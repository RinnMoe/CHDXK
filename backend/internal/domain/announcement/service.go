package announcement

import "context"

type Service struct {
	repo AnnouncementRepository
}

func NewService(repo AnnouncementRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input SaveInput) (*Announcement, error) {
	item, err := New(input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, id int, input SaveInput) (*Announcement, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrAnnouncementNotFound
	}
	if err := item.Update(input); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrAnnouncementNotFound
	}
	return nil
}
