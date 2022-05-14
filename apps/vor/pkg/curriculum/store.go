package curriculum

import (
	"github.com/chrsep/vor/pkg/domain"
	"github.com/google/uuid"
)

type Store interface {
	GetArea(areaId string) (*domain.Area, error)
	GetAreaSubjects(areaId string) ([]domain.Subject, error)
	GetSubjectMaterials(subjectId string) ([]domain.Material, error)
	GetMaterial(materialId string) (*domain.Material, error)
	NewArea(curriculumId string, name string, description string) (*domain.Area, error)
	NewSubject(name string, areaId string, materials []domain.Material, description string) (*domain.Subject, error)
	NewMaterial(subjectId string, name string, description string) (*domain.Material, error)
	GetSubject(id string) (*domain.Subject, error)
	UpdateMaterial(id string, name *string, order *int, description *string, subjectId *uuid.UUID) error
	DeleteArea(id string) error
	DeleteSubject(id string) error
	ReplaceSubject(subject domain.Subject) error
	UpdateArea(areaId string, name string) error
	CheckSubjectPermissions(subjectId string, userId string) (bool, error)
	CheckAreaPermissions(subjectId string, userId string) (bool, error)
	CheckCurriculumPermission(curriculumId string, userId string) (bool, error)
	CheckMaterialPermission(materialId string, userId string) (bool, error)
	UpdateCurriculum(curriculumId string, name *string, description *string) (*domain.Curriculum, error)
	UpdateSubject(id string, name *string, order *int, description *string, areaId *uuid.UUID) (*domain.Subject, error)
	DeleteMaterial(id string) error
}
