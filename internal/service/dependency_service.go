package service

import (
	"akasha/internal/domain"
	"akasha/internal/repository"
	"time"

	"github.com/Masterminds/semver/v3"
)

type DependencyService struct {
	depRepo *repository.DependencyRepo
	branchRepo *repository.BranchRepo
}

func NewDependencyService(depRepo *repository.DependencyRepo, branchRepo *repository.BranchRepo) *DependencyService {
	return &DependencyService{
		depRepo: depRepo,
		branchRepo: branchRepo,
	}
}

func (s *DependencyService) List(branch string) ([]domain.Dependency, error) {
	return s.depRepo.FindLatestByBranch(branch)
}

func (s *DependencyService) Get(name, branch string) (*domain.Dependency, error) {
	return s.depRepo.FindByNameAndBranch(name, branch)
}

func (s *DependencyService) History(name, branch string) ([]domain.Dependency, error) {
	return s.depRepo.FindHistory(name, branch)
}

func (s *DependencyService) GetAt(name, branch string, at time.Time) (*domain.Dependency, error) {
	return s.depRepo.FindAt(name, branch, at)
}

// HistoryBetween 获取某时间段内的依赖变更
func (s *DependencyService) HistoryBetween(name, branch string, startAt, endAt time.Time) ([]domain.Dependency, error) {
	return s.depRepo.FindHistoryBetween(name, branch, startAt, endAt)
}

func (s *DependencyService) CreateOrUpdate(d *domain.Dependency) error {
	existing, err := s.depRepo.FindByNameAndBranch(d.Name, d.Branch)
	if err == nil && existing != nil {
		d.ID = 0
		d.CreatedAt = time.Now()
		return s.depRepo.Create(d)
	}
	d.CreatedAt = time.Now()
	return s.depRepo.Create(d)
}

func (s *DependencyService) Delete(name, branch string) error {
	branchInfo, err := s.branchRepo.FindByName(branch)
	if err != nil {
		return err
	}
	if branchInfo.Status == domain.StatusActive {
		return &ErrForbidden{msg: "只能在创建中的分支上删除条目"}
	}
	return s.depRepo.Delete(name, branch)
}

type ErrForbidden struct {
	msg string
}

func (e *ErrForbidden) Error() string {
	return e.msg
}

func CompareVersion(v1, v2 string) int {
	s1, err1 := semver.NewVersion(v1)
	s2, err2 := semver.NewVersion(v2)
	if err1 != nil || err2 != nil {
		if v1 == v2 {
			return 0
		}
		if v1 > v2 {
			return 1
		}
		return -1
	}
	return s1.Compare(s2)
}

// GetDepsAt 批量闪回查询 - 获取某分支在某时间点的所有依赖
func (s *DependencyService) GetDepsAt(branch string, at time.Time) ([]domain.Dependency, error) {
	return s.depRepo.FindDepsAt(branch, at)
}