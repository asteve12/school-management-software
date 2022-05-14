package student

import (
	"github.com/chrsep/vor/pkg/domain"
	"github.com/chrsep/vor/pkg/postgres"
	"github.com/google/uuid"
	"mime/multipart"
	"time"
)

type Store interface {
	InsertObservation(studentId string, creatorId string, longDesc string, shortDesc string, category string, eventTime time.Time, images []uuid.UUID, areaId uuid.UUID, visibleToGuardians bool) (*postgres.Observation, error)
	GetObservations(studentId string, search string, startDate string, endDate string) ([]postgres.Observation, error)
	GetProgress(studentId string) ([]postgres.StudentMaterialProgress, error)
	UpdateProgress(progress postgres.StudentMaterialProgress) (*postgres.StudentMaterialProgress, error)
	Get(studentId string) (*postgres.Student, error)
	UpdateStudent(student *postgres.Student) error
	DeleteStudent(studentId string) error
	CheckPermissions(studentId string, userId string) (bool, error)
	InsertAttendance(studentId string, classId string, date time.Time) (*postgres.Attendance, error)
	GetAttendance(studentId string) ([]postgres.Attendance, error)
	InsertGuardianRelation(studentId string, guardianId string, relationship int) error
	DeleteGuardianRelation(studentId string, guardianId string) error
	GetGuardianRelation(studentId string, guardianId string) (*postgres.GuardianToStudent, error)
	NewClassRelation(studentId string, classId string) error
	DeleteClassRelation(studentId string, classId string) error
	GetLessonPlans(studentId string, date time.Time) ([]postgres.LessonPlan, error)
	CreateImage(studentId string, file multipart.File, header *multipart.FileHeader) (domain.Image, error)
	FindStudentImages(id string) ([]postgres.Image, error)
	FindStudentVideos(studentId string) ([]domain.Video, error)
	FindCurriculum(studentId string) (domain.Curriculum, error)
}
