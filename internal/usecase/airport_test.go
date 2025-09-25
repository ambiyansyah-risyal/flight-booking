package usecase

import (
    "context"
    "testing"

    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

type fakeAirportRepo struct{
    created []*domain.Airport
    createErr error
    list []domain.Airport
    lastLimit int
    lastOffset int
    updateErr error
    deleteErr error
}

func (f *fakeAirportRepo) Create(ctx context.Context, a *domain.Airport) error {
    if f.createErr != nil { return f.createErr }
    f.created = append(f.created, &domain.Airport{Code: a.Code, City: a.City})
    return nil
}

func (f *fakeAirportRepo) GetByCode(ctx context.Context, code string) (*domain.Airport, error) {
    for _, a := range f.list {
        if a.Code == code { return &a, nil }
    }
    return nil, domain.ErrAirportNotFound
}

func (f *fakeAirportRepo) List(ctx context.Context, limit, offset int) ([]domain.Airport, error) {
    f.lastLimit, f.lastOffset = limit, offset
    return f.list, nil
}

func (f *fakeAirportRepo) Update(ctx context.Context, code string, city string) error {
    if f.updateErr != nil { return f.updateErr }
    for i := range f.list {
        if f.list[i].Code == code { f.list[i].City = city; return nil }
    }
    return domain.ErrAirportNotFound
}

func (f *fakeAirportRepo) Delete(ctx context.Context, code string) error {
    if f.deleteErr != nil { return f.deleteErr }
    for i := range f.list {
        if f.list[i].Code == code { f.list = append(f.list[:i], f.list[i+1:]...); return nil }
    }
    return domain.ErrAirportNotFound
}

func TestAirportUsecase_Create_ValidatesAndCallsRepo(t *testing.T) {
    repo := &fakeAirportRepo{}
    uc := NewAirportUsecase(repo)

    got, err := uc.Create(context.Background(), "cgk", " Jakarta ")
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if got.Code != "CGK" || got.City != "Jakarta" {
        t.Fatalf("normalized mismatch: %+v", got)
    }
    if len(repo.created) != 1 || repo.created[0].Code != "CGK" {
        t.Fatalf("repo not called or wrong args: %+v", repo.created)
    }
}

func TestAirportUsecase_Create_Invalid(t *testing.T) {
    repo := &fakeAirportRepo{}
    uc := NewAirportUsecase(repo)
    if _, err := uc.Create(context.Background(), "", "City"); err != domain.ErrInvalidAirportCode {
        t.Fatalf("want ErrInvalidAirportCode, got %v", err)
    }
    if len(repo.created) != 0 { t.Fatalf("repo should not be called on invalid input") }
}

func TestAirportUsecase_List_Pagination(t *testing.T) {
    repo := &fakeAirportRepo{list: []domain.Airport{{Code:"CGK", City:"Jakarta"}}}
    uc := NewAirportUsecase(repo)
    items, err := uc.List(context.Background(), 10, 5)
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if len(items) != 1 { t.Fatalf("want 1 item, got %d", len(items)) }
    if repo.lastLimit != 10 || repo.lastOffset != 5 {
        t.Fatalf("unexpected limit/offset: %d/%d", repo.lastLimit, repo.lastOffset)
    }
}

func TestAirportUsecase_Update_Delete(t *testing.T) {
    repo := &fakeAirportRepo{list: []domain.Airport{{Code:"CGK", City:"Jakarta"}}}
    uc := NewAirportUsecase(repo)
    if err := uc.Update(context.Background(), "cgk", "NewCity"); err != nil {
        t.Fatalf("update err: %v", err)
    }
    if repo.list[0].City != "NewCity" { t.Fatalf("city not updated: %+v", repo.list[0]) }
    if err := uc.Delete(context.Background(), "cgk"); err != nil { t.Fatalf("delete err: %v", err) }
    if len(repo.list) != 0 { t.Fatalf("item not deleted") }
}

func TestAirportUsecase_List_Defaults(t *testing.T) {
    repo := &fakeAirportRepo{list: []domain.Airport{}}
    uc := NewAirportUsecase(repo)
    if _, err := uc.List(context.Background(), 9999, -1); err != nil {
        t.Fatalf("unexpected err: %v", err)
    }
    if repo.lastLimit != 50 || repo.lastOffset != 0 {
        t.Fatalf("defaults not applied: limit=%d offset=%d", repo.lastLimit, repo.lastOffset)
    }
}

func TestAirportUsecase_Create_RepoError(t *testing.T) {
    repo := &fakeAirportRepo{createErr: domain.ErrAirportExists}
    uc := NewAirportUsecase(repo)
    if _, err := uc.Create(context.Background(), "CGK", "Jakarta"); err != domain.ErrAirportExists {
        t.Fatalf("want ErrAirportExists, got %v", err)
    }
}

func TestAirportUsecase_Update_Invalid(t *testing.T) {
    repo := &fakeAirportRepo{}
    uc := NewAirportUsecase(repo)
    if err := uc.Update(context.Background(), "", "City"); err != domain.ErrInvalidAirportCode {
        t.Fatalf("expected code validation error, got %v", err)
    }
}

func TestAirportUsecase_Delete_Invalid(t *testing.T) {
    repo := &fakeAirportRepo{}
    uc := NewAirportUsecase(repo)
    if err := uc.Delete(context.Background(), ""); err != domain.ErrInvalidAirportCode {
        t.Fatalf("expected code validation error, got %v", err)
    }
}
