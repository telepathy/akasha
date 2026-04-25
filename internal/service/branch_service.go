package service

import (
	"akasha/internal/domain"
	"akasha/internal/repository"
)

type BranchService struct {
	branchRepo *repository.BranchRepo
	depRepo    *repository.DependencyRepo
}

func NewBranchService(branchRepo *repository.BranchRepo, depRepo *repository.DependencyRepo) *BranchService {
	return &BranchService{
		branchRepo: branchRepo,
		depRepo:    depRepo,
	}
}

func (s *BranchService) List() ([]domain.Branch, error) {
	return s.branchRepo.FindAll()
}

func (s *BranchService) Get(name string) (*domain.Branch, error) {
	return s.branchRepo.FindByName(name)
}

func (s *BranchService) Create(name, baseBranch string) error {
	if s.branchRepo.Exists(name) {
		return &ErrBranchExists{name: name}
	}
	return s.branchRepo.CreateBranch(name, baseBranch)
}

func (s *BranchService) Delete(name string) error {
	// Soft delete: mark branch as deleted
	return s.branchRepo.UpdateStatus(name, "deleted")
}

func (s *BranchService) IsDeleted(name string) bool {
	branch, err := s.branchRepo.FindByName(name)
	if err != nil {
		return true
	}
	return branch.Status == "deleted"
}

func (s *BranchService) Archive(name string) error {
	// Archive: mark branch as archived (locked), content frozen but still accessible
	return s.branchRepo.UpdateStatus(name, "archived")
}

func (s *BranchService) Unlock(name string) error {
	// Unlock: change archived branch back to active
	return s.branchRepo.UpdateStatus(name, "active")
}

func (s *BranchService) IsArchived(name string) bool {
	branch, err := s.branchRepo.FindByName(name)
	if err != nil {
		return false
	}
	return branch.Status == "archived"
}

func (s *BranchService) CanModify(name string) bool {
	branch, err := s.branchRepo.FindByName(name)
	if err != nil {
		return false
	}
	// Only active branches can be modified
	return branch.Status == "active"
}

func (s *BranchService) Merge(fromBranch, toBranch string) error {
	fromDeps, err := s.depRepo.FindLatestByBranch(fromBranch)
	if err != nil {
		return err
	}

	for _, fromDep := range fromDeps {
		toDep, err := s.depRepo.FindByNameAndBranch(fromDep.Name, toBranch)
		if err != nil {
			continue
		}
		if CompareVersion(fromDep.Version, toDep.Version) > 0 {
			fromDep.ID = 0
			fromDep.Branch = toBranch
			if err := s.depRepo.Create(&fromDep); err != nil {
				return err
			}
		}
	}
	return nil
}

// History 获取分支的历史变更记录（基于依赖的历史）
func (s *BranchService) History(name string) (map[string][]domain.Dependency, error) {
	deps, err := s.depRepo.FindLatestByBranch(name)
	if err != nil {
		return nil, err
	}
	// 返回每个依赖的历史
	result := make(map[string][]domain.Dependency)
	for _, dep := range deps {
		history, err := s.depRepo.FindHistory(dep.Name, name)
		if err != nil {
			continue
		}
		result[dep.Name] = history
	}
	return result, nil
}

type ErrBranchExists struct {
	name string
}

func (e *ErrBranchExists) Error() string {
	return "分支已存在: " + e.name
}