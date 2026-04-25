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
	return s.branchRepo.UpdateStatus(name, "archived")
}

func (s *BranchService) Unlock(name string) error {
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
	return branch.Status == "active"
}

type MergeStrategy string

const (
	MergeKeepHigher   MergeStrategy = "keep_higher"
	MergeForceSource  MergeStrategy = "force_source"
	MergeForceTarget  MergeStrategy = "force_target"
)

type MergeConfig struct {
	SourceBranch string        `json:"sourceBranch"`
	TargetBranch string        `json:"targetBranch"`
	Strategy     MergeStrategy `json:"strategy"`
	AddMissing   bool          `json:"addMissing"`
	DryRun       bool          `json:"dryRun"`
}

type MergeConflict struct {
	Name        string `json:"name"`
	SourceCoord string `json:"sourceCoord"`
	TargetCoord string `json:"targetCoord"`
	SourceVer   string `json:"sourceVersion"`
	TargetVer   string `json:"targetVersion"`
	Reason      string `json:"reason"`
}

type MergeResult struct {
	Added     int            `json:"added"`
	Updated   int            `json:"updated"`
	Skipped   int            `json:"skipped"`
	Conflicts []MergeConflict `json:"conflicts"`
	Details   []string       `json:"details"`
}

func (s *BranchService) PreviewMerge(config MergeConfig) (*MergeResult, error) {
	return s.executeMerge(config, true)
}

func (s *BranchService) ExecuteMerge(config MergeConfig) (*MergeResult, error) {
	return s.executeMerge(config, false)
}

func (s *BranchService) executeMerge(config MergeConfig, dryRun bool) (*MergeResult, error) {
	sourceDeps, err := s.depRepo.FindLatestByBranch(config.SourceBranch)
	if err != nil {
		return nil, err
	}

	targetDeps, err := s.depRepo.FindLatestByBranch(config.TargetBranch)
	if err != nil {
		return nil, err
	}

	sourceMap := make(map[string]domain.Dependency)
	for _, dep := range sourceDeps {
		sourceMap[dep.Name] = dep
	}

	targetMap := make(map[string]domain.Dependency)
	for _, dep := range targetDeps {
		targetMap[dep.Name] = dep
	}

	result := &MergeResult{}

	for name, sourceDep := range sourceMap {
		targetDep, exists := targetMap[name]

		if !exists {
			if config.AddMissing {
				result.Added++
				result.Details = append(result.Details, "添加: "+name+" "+sourceDep.Version)
				if !dryRun {
					newDep := sourceDep
					newDep.ID = 0
					newDep.Branch = config.TargetBranch
					if err := s.depRepo.Create(&newDep); err != nil {
						return nil, err
					}
				}
			} else {
				result.Skipped++
				result.Details = append(result.Details, "跳过(目标无此依赖): "+name)
			}
			continue
		}

		if sourceDep.Version == targetDep.Version {
			result.Skipped++
			continue
		}

		cmp := CompareVersion(sourceDep.Version, targetDep.Version)

		switch config.Strategy {
		case MergeForceSource:
			result.Updated++
			result.Details = append(result.Details, "更新(强制源): "+name+" "+targetDep.Version+" -> "+sourceDep.Version)
			if !dryRun {
				newDep := sourceDep
				newDep.ID = 0
				newDep.Branch = config.TargetBranch
				if err := s.depRepo.Create(&newDep); err != nil {
					return nil, err
				}
			}

		case MergeForceTarget:
			result.Skipped++
			result.Details = append(result.Details, "跳过(保留目标): "+name)

		case MergeKeepHigher:
			if cmp > 0 {
				result.Updated++
				result.Details = append(result.Details, "更新(源更高): "+name+" "+targetDep.Version+" -> "+sourceDep.Version)
				if !dryRun {
					newDep := sourceDep
					newDep.ID = 0
					newDep.Branch = config.TargetBranch
					if err := s.depRepo.Create(&newDep); err != nil {
						return nil, err
					}
				}
			} else {
				result.Conflicts = append(result.Conflicts, MergeConflict{
					Name:        name,
					SourceCoord: sourceDep.MavenCoord(),
					TargetCoord: targetDep.MavenCoord(),
					SourceVer:   sourceDep.Version,
					TargetVer:   targetDep.Version,
					Reason:      "目标版本更高",
				})
			}

		default:
			result.Skipped++
		}
	}

	for name, targetDep := range targetMap {
		if _, exists := sourceMap[name]; !exists {
			result.Conflicts = append(result.Conflicts, MergeConflict{
				Name:        name,
				SourceCoord: "",
				TargetCoord: targetDep.MavenCoord(),
				SourceVer:   "",
				TargetVer:   targetDep.Version,
				Reason:      "源分支无此依赖，可能被误删",
			})
		}
	}

	return result, nil
}

func (s *BranchService) History(name string) (map[string][]domain.Dependency, error) {
	deps, err := s.depRepo.FindLatestByBranch(name)
	if err != nil {
		return nil, err
	}
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