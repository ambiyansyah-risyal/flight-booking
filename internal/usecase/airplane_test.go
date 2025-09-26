package usecase

import (
    "context"
    "testing"

    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

type fakeAirplaneRepo struct{
    list []domain.Airplane
    createErr error
    updateErr error
    deleteErr error
}

func (f *fakeAirplaneRepo) Create(ctx context.Context, a *domain.Airplane) error {
    if f.createErr != nil { return f.createErr }
    f.list = append(f.list, domain.Airplane{Code:a.Code, SeatCapacity:a.SeatCapacity})
    return nil
}
func (f *fakeAirplaneRepo) GetByCode(ctx context.Context, code string) (*domain.Airplane, error) { return nil, nil }
func (f *fakeAirplaneRepo) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) { return f.list, nil }
func (f *fakeAirplaneRepo) UpdateSeats(ctx context.Context, code string, seats int) error { if f.updateErr!=nil {return f.updateErr}; return nil }
func (f *fakeAirplaneRepo) Delete(ctx context.Context, code string) error { if f.deleteErr!=nil {return f.deleteErr}; return nil }

func TestAirplaneUsecase_Create_List(t *testing.T) {
    r := &fakeAirplaneRepo{}
    uc := NewAirplaneUsecase(r)
    if _, err := uc.Create(context.Background(), " ab ", 10); err != nil { t.Fatalf("create: %v", err) }
    items, err := uc.List(context.Background(), 1000, -1)
    if err != nil || len(items) != 1 { t.Fatalf("list: %v n=%d", err, len(items)) }
}

func TestAirplaneUsecase_Update_Delete_Validate(t *testing.T) {
    r := &fakeAirplaneRepo{}
    uc := NewAirplaneUsecase(r)
    if err := uc.UpdateSeats(context.Background(), "", 1); err != domain.ErrInvalidAirplaneCode { t.Fatalf("want code err: %v", err) }
    if err := uc.UpdateSeats(context.Background(), "OK", 0); err != domain.ErrInvalidSeatCapacity { t.Fatalf("want seat err: %v", err) }
    if err := uc.Delete(context.Background(), ""); err != domain.ErrInvalidAirplaneCode { t.Fatalf("want code err: %v", err) }
}

func TestAirplaneUsecase_Update_Delete_Success(t *testing.T) {
    r := &fakeAirplaneRepo{}
    uc := NewAirplaneUsecase(r)
    
    // Test successful update
    if err := uc.UpdateSeats(context.Background(), "TEST", 150); err != nil {
        t.Fatalf("update seats failed: %v", err)
    }
    
    // Test successful delete
    if err := uc.Delete(context.Background(), "TEST"); err != nil {
        t.Fatalf("delete failed: %v", err)
    }
}

func TestAirplaneUsecase_Update_Delete_RepoErrors(t *testing.T) {
    r := &fakeAirplaneRepo{updateErr: domain.ErrAirplaneNotFound, deleteErr: domain.ErrAirplaneNotFound}
    uc := NewAirplaneUsecase(r)
    
    // Test repo error for update
    if err := uc.UpdateSeats(context.Background(), "TEST", 150); err != domain.ErrAirplaneNotFound {
        t.Fatalf("want repo error: %v", err)
    }
    
    // Test repo error for delete
    if err := uc.Delete(context.Background(), "TEST"); err != domain.ErrAirplaneNotFound {
        t.Fatalf("want repo error: %v", err)
    }
}
