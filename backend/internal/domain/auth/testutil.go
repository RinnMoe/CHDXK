//go:build test

package auth

import "context"

type MockUserRepository struct {
	Users         map[int]*User
	FindByIDCalls int

	OnUpdate    func(context.Context, *User) error
	OnFindByID  func(context.Context, int) (*User, error)
	OnFindAdmin func(context.Context) ([]User, error)
}

func NewMockUserRepository(users map[int]*User) *MockUserRepository {
	if users == nil {
		users = map[int]*User{}
	}
	return &MockUserRepository{Users: users}
}

func (r *MockUserRepository) Update(ctx context.Context, u *User) error {
	if r.OnUpdate != nil {
		return r.OnUpdate(ctx, u)
	}
	r.ensureMap()
	copy := *u
	r.Users[u.ID] = &copy
	return nil
}

func (r *MockUserRepository) FindByID(ctx context.Context, id int) (*User, error) {
	if r.OnFindByID != nil {
		return r.OnFindByID(ctx, id)
	}
	r.FindByIDCalls++
	r.ensureMap()
	u, ok := r.Users[id]
	if !ok {
		return nil, nil
	}
	copy := *u
	return &copy, nil
}

func (r *MockUserRepository) FindAdmin(ctx context.Context) ([]User, error) {
	if r.OnFindAdmin != nil {
		return r.OnFindAdmin(ctx)
	}
	r.ensureMap()
	users := make([]User, 0)
	for _, u := range r.Users {
		if u.IsAdmin() {
			users = append(users, *u)
		}
	}
	return users, nil
}

func (r *MockUserRepository) ensureMap() {
	if r.Users == nil {
		r.Users = map[int]*User{}
	}
}

type MockApiKeyRepository struct {
	Key          *ApiKey
	KeyByID      map[int64]*ApiKey
	Keys         []ApiKey
	SystemKeys   []ApiKey
	Count        int
	Created      *ApiKey
	Deleted      bool
	Touched      bool
	DeleteOK     bool
	DeleteID     int64
	GetByIDCalls int

	OnGetByID     func(context.Context, int64) (*ApiKey, error)
	OnListSystem  func(context.Context) ([]ApiKey, error)
	OnListByUser  func(context.Context, int) ([]ApiKey, error)
	OnCountByUser func(context.Context, int) (int, error)
	OnCreate      func(context.Context, *ApiKey) error
	OnDelete      func(context.Context, int64) (bool, error)
}

func (r *MockApiKeyRepository) GetByID(ctx context.Context, id int64) (*ApiKey, error) {
	if r.OnGetByID != nil {
		return r.OnGetByID(ctx, id)
	}
	r.GetByIDCalls++
	if r.Key != nil && r.Key.ID == id {
		copy := *r.Key
		return &copy, nil
	}
	if r.KeyByID != nil {
		if key, ok := r.KeyByID[id]; ok {
			copy := *key
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *MockApiKeyRepository) ListByUser(ctx context.Context, userID int) ([]ApiKey, error) {
	if r.OnListByUser != nil {
		return r.OnListByUser(ctx, userID)
	}
	return r.Keys, nil
}

func (r *MockApiKeyRepository) ListSystem(ctx context.Context) ([]ApiKey, error) {
	if r.OnListSystem != nil {
		return r.OnListSystem(ctx)
	}
	return r.SystemKeys, nil
}

func (r *MockApiKeyRepository) CountByUser(ctx context.Context, userID int) (int, error) {
	if r.OnCountByUser != nil {
		return r.OnCountByUser(ctx, userID)
	}
	return r.Count, nil
}

func (r *MockApiKeyRepository) Create(ctx context.Context, apiKey *ApiKey) error {
	if r.OnCreate != nil {
		return r.OnCreate(ctx, apiKey)
	}
	copy := *apiKey
	r.Created = &copy
	return nil
}

func (r *MockApiKeyRepository) Delete(ctx context.Context, id int64) (bool, error) {
	if r.OnDelete != nil {
		return r.OnDelete(ctx, id)
	}
	r.DeleteID = id
	r.Deleted = r.DeleteOK
	return r.DeleteOK, nil
}
