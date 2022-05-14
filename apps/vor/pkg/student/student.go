package student

import (
	"github.com/chrsep/vor/pkg/domain"
	"github.com/chrsep/vor/pkg/imgproxy"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/chrsep/vor/pkg/auth"
	"github.com/chrsep/vor/pkg/postgres"
	"github.com/chrsep/vor/pkg/rest"
	"github.com/go-chi/chi"
	richErrors "github.com/pkg/errors"
)

func NewRouter(s rest.Server, store Store) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/{studentId}", func(r chi.Router) {
		r.Use(authorizationMiddleware(s, store))
		r.Method("GET", "/", getStudent(s, store))
		r.Method("DELETE", "/", deleteStudent(s, store))
		r.Method("PATCH", "/", patchStudent(s, store))

		r.Method("POST", "/observations", postObservation(s, store))
		r.Method("GET", "/observations", getObservation(s, store))

		r.Method("POST", "/attendances", postAttendance(s, store))
		r.Method("GET", "/attendances", getAttendance(s, store))

		r.Method("POST", "/guardianRelations", postNewGuardianRelation(s, store))
		r.Method("DELETE", "/guardianRelations/{guardianId}", deleteGuardianRelation(s, store))

		r.Method("POST", "/classes", postClassRelation(s, store))
		r.Method("DELETE", "/classes", deleteClassRelation(s, store))

		r.Method("GET", "/plans", getPlans(s, store))

		r.Method("POST", "/images", postNewImage(s, store))
		r.Method("GET", "/images", getStudentImages(s, store))

		r.Method("GET", "/videos", getStudentVideos(s, store))

		r.Route("/materialsProgress", func(r chi.Router) {
			r.Method("GET", "/", getMaterialProgress(s, store))
			r.Method("PATCH", "/{materialId}", upsertMaterialProgress(s, store))
			r.Method("GET", "/export/pdf", exportMaterialProgressPdf(s, store))
			r.Method("GET", "/export/csv", exportMaterialProgressCsv(s, store))
		})

	})
	return r
}

func postClassRelation(s rest.Server, store Store) http.Handler {
	type reqBody struct {
		ClassId string `json:"classId"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		var body reqBody
		if err := rest.ParseJson(r.Body, &body); err != nil {
			return rest.NewParseJsonError(err)
		}

		if err := store.NewClassRelation(studentId, body.ClassId); err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed to create class relationship",
				Error:   err,
			}
		}
		return nil
	})
}

func deleteClassRelation(s rest.Server, store Store) http.Handler {
	type reqBody struct {
		ClassId string `json:"classId"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		var body reqBody
		if err := rest.ParseJson(r.Body, &body); err != nil {
			return rest.NewParseJsonError(err)
		}

		if err := store.DeleteClassRelation(studentId, body.ClassId); err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed to delete class relationship",
				Error:   err,
			}
		}
		return nil
	})
}

func authorizationMiddleware(s rest.Server, store Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
			studentId := chi.URLParam(r, "studentId")

			// Verify user access to the school
			session, ok := auth.GetSessionFromCtx(r.Context())
			if !ok {
				return auth.NewGetSessionError()
			}

			// Check if user is related to the school
			userHasAccess, err := store.CheckPermissions(studentId, session.UserId)
			if err != nil {
				return &rest.Error{
					Code:    http.StatusInternalServerError,
					Message: "Internal Server Error",
					Error:   err,
				}
			}
			if !userHasAccess {
				return &rest.Error{
					Code:    http.StatusNotFound,
					Message: "We can't find the specified student",
					Error:   err,
				}
			}

			next.ServeHTTP(w, r)
			return nil
		})
	}
}
func postAttendance(s rest.Server, store Store) http.Handler {
	type requestBody struct {
		StudentId string    `json:"studentId"`
		ClassId   string    `json:"classId"`
		Date      time.Time `json:"date"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		var requestBody requestBody
		if err := rest.ParseJson(r.Body, &requestBody); err != nil {
			return rest.NewParseJsonError(err)
		}
		attendance, err := store.InsertAttendance(requestBody.StudentId, requestBody.ClassId, requestBody.Date)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusNotFound,
				Message: "Can't create attendance",
				Error:   err,
			}
		}
		if err := rest.WriteJson(w, attendance); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}
func getAttendance(s rest.Server, store Store) http.Handler {

	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		id := chi.URLParam(r, "studentId")
		attendance, err := store.GetAttendance(id)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusNotFound,
				Message: "Can't find attendance",
				Error:   err,
			}
		}
		if err := rest.WriteJson(w, attendance); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func postNewGuardianRelation(s rest.Server, store Store) http.Handler {
	type requestBody struct {
		Id           string `json:"id"`
		Relationship int    `json:"relationship"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		var body requestBody
		if err := rest.ParseJson(r.Body, &body); err != nil {
			return rest.NewParseJsonError(err)
		}

		if err := store.InsertGuardianRelation(studentId, body.Id, body.Relationship); err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed to save relationship",
				Error:   err,
			}
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	})
}

func deleteGuardianRelation(s rest.Server, store Store) http.Handler {
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")
		guardianId := chi.URLParam(r, "guardianId")

		if err := store.DeleteGuardianRelation(studentId, guardianId); err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed to delete relationship",
				Error:   err,
			}
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

func getStudent(s rest.Server, store Store) http.Handler {
	type Guardian struct {
		Id           string `json:"id"`
		Name         string `json:"name"`
		Relationship int    `json:"relationship"`
		Email        string `json:"email"`
	}
	type Class struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type responseBody struct {
		Id          string     `json:"id"`
		Name        string     `json:"name"`
		DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
		DateOfEntry *time.Time `json:"dateOfEntry,omitempty"`
		Gender      int        `json:"gender"`
		Note        string     `json:"note"`
		CustomId    string     `json:"customId"`
		Active      bool       `json:"active"`
		ProfilePic  string     `json:"profilePic"`
		Classes     []Class    `json:"classes"`
		Guardians   []Guardian `json:"guardians"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		id := chi.URLParam(r, "studentId")

		student, err := store.Get(id)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusNotFound,
				Message: "Can't find student with specified id",
				Error:   err,
			}
		}

		guardians := make([]Guardian, len(student.Guardians))
		for i, guardian := range student.Guardians {
			relation, err := store.GetGuardianRelation(id, guardian.Id)
			if err != nil {
				return &rest.Error{
					Code:    http.StatusInternalServerError,
					Message: "can't find student to guardian relation",
					Error:   err,
				}
			}
			guardians[i] = Guardian{
				Id:           guardian.Id,
				Name:         guardian.Name,
				Relationship: int(relation.Relationship),
				Email:        guardian.Email,
			}
		}
		classes := make([]Class, len(student.Classes))
		for i, class := range student.Classes {
			classes[i] = Class{
				Id:   class.Id,
				Name: class.Name,
			}
		}
		response := responseBody{
			Id:          student.Id,
			Name:        student.Name,
			Gender:      int(student.Gender),
			CustomId:    student.CustomId,
			Note:        student.Note,
			DateOfBirth: student.DateOfBirth,
			DateOfEntry: student.DateOfEntry,
			Guardians:   guardians,
			Classes:     classes,
			Active:      *student.Active,
		}
		if student.ProfileImage.ObjectKey != "" {
			response.ProfilePic = imgproxy.GenerateUrlFromS3(student.ProfileImage.ObjectKey, 80, 80)
		}
		if err := rest.WriteJson(w, response); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func deleteStudent(s rest.Server, store Store) http.Handler {
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId") // from a route like /users/{userID}
		if err := store.DeleteStudent(studentId); err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "Failed deleting student",
				Error:   err,
			}
		}
		return nil
	})
}

func patchStudent(s rest.Server, store Store) http.Handler {
	type requestBody struct {
		Name           string          `json:"name"`
		DateOfBirth    *time.Time      `json:"dateOfBirth"`
		DateOfEntry    *time.Time      `json:"dateOfEntry"`
		CustomId       string          `json:"customId"`
		Gender         postgres.Gender `json:"gender"`
		Active         *bool           `json:"active"`
		ProfileImageId string          `json:"profileImageId"`
		Note           string          `json:"note"`
	}
	type responseBody struct {
		Id          string     `json:"id"`
		Name        string     `json:"name"`
		DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
		Active      *bool      `json:"active"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		targetId := chi.URLParam(r, "studentId") // from a route like /users/{userID}

		var requestBody requestBody
		if err := rest.ParseJson(r.Body, &requestBody); err != nil {
			return rest.NewParseJsonError(err)
		}

		oldStudent, err := store.Get(targetId)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusNotFound,
				Message: "Can't find old student data",
				Error:   err,
			}
		}

		newStudent := oldStudent
		newStudent.Name = requestBody.Name
		newStudent.DateOfBirth = requestBody.DateOfBirth
		newStudent.Gender = requestBody.Gender
		newStudent.CustomId = requestBody.CustomId
		newStudent.DateOfEntry = requestBody.DateOfEntry
		newStudent.Active = requestBody.Active
		newStudent.ProfileImageId = requestBody.ProfileImageId
		newStudent.Note = requestBody.Note
		if err := store.UpdateStudent(newStudent); err != nil {
			return &rest.Error{Code: http.StatusInternalServerError, Message: "Failed updating old student data", Error: err}
		}

		response := responseBody{
			Id:          newStudent.Id,
			Name:        newStudent.Name,
			DateOfBirth: newStudent.DateOfBirth,
			Active:      newStudent.Active,
		}
		if err := rest.WriteJson(w, response); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func postObservation(s rest.Server, store Store) http.Handler {
	type requestBody struct {
		ShortDesc          string      `json:"shortDesc"`
		LongDesc           string      `json:"longDesc"`
		CategoryId         string      `json:"categoryId"`
		EventTime          *time.Time  `json:"eventTime"`
		Images             []uuid.UUID `json:"images"`
		AreaId             uuid.UUID   `json:"areaId"`
		VisibleToGuardians bool        `json:"visibleToGuardians"`
	}

	type area struct {
		Id   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	type image struct {
		Id           uuid.UUID `json:"id"`
		ThumbnailUrl string    `json:"thumbnailUrl"`
		OriginalUrl  string    `json:"originalUrl"`
	}
	type responseBody struct {
		Id                 string    `json:"id"`
		ShortDesc          string    `json:"shortDesc"`
		LongDesc           string    `json:"longDesc"`
		CategoryId         string    `json:"categoryId"`
		CreatedDate        time.Time `json:"createdDate"`
		EventTime          time.Time `json:"eventTime"`
		Images             []image   `json:"images"`
		Area               *area     `json:"area,omitempty"`
		CreatorId          string    `json:"creatorId,omitempty"`
		CreatorName        string    `json:"creatorName,omitempty"`
		VisibleToGuardians bool      `json:"visibleToGuardians"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		id := chi.URLParam(r, "studentId")
		session, ok := auth.GetSessionFromCtx(r.Context())
		if !ok {
			return &rest.Error{
				Code:    http.StatusUnauthorized,
				Message: "You don't have access to this student",
				Error:   richErrors.New("user is not authorized to add observation."),
			}
		}

		var body requestBody
		if err := rest.ParseJson(r.Body, &body); err != nil {
			return rest.NewParseJsonError(err)
		}

		var eventTime = time.Now()
		if body.EventTime != nil {
			eventTime = *body.EventTime
		}
		observation, err := store.InsertObservation(id,
			session.UserId,
			body.LongDesc,
			body.ShortDesc,
			body.CategoryId,
			eventTime,
			body.Images,
			body.AreaId,
			body.VisibleToGuardians,
		)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "Failed inserting observation",
				Error:   err,
			}
		}

		images := make([]image, 0)
		for i := range observation.Images {
			images = append(images, image{
				Id:           observation.Images[i].Id,
				ThumbnailUrl: imgproxy.GenerateUrlFromS3(observation.Images[i].ObjectKey, 80, 80),
				OriginalUrl:  imgproxy.GenerateOriginalUrlFromS3(observation.Images[i].ObjectKey),
			})
		}

		response := responseBody{
			Id:                 observation.Id,
			ShortDesc:          observation.ShortDesc,
			LongDesc:           observation.LongDesc,
			CategoryId:         observation.CategoryId,
			EventTime:          observation.EventTime,
			Images:             images,
			CreatorId:          observation.CreatorId,
			CreatorName:        observation.Creator.Name,
			CreatedDate:        observation.CreatedDate,
			VisibleToGuardians: observation.VisibleToGuardians,
		}
		if observation.AreaId != uuid.Nil {
			response.Area = &area{
				Id:   observation.AreaId,
				Name: observation.Area.Name,
			}
		}
		w.WriteHeader(http.StatusCreated)
		if err := rest.WriteJson(w, response); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func getObservation(s rest.Server, store Store) http.Handler {
	type image struct {
		Id           uuid.UUID `json:"id"`
		ThumbnailUrl string    `json:"thumbnailUrl"`
		OriginalUrl  string    `json:"originalUrl"`
	}
	type area struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type responseBody struct {
		Id                 string    `json:"id"`
		StudentName        string    `json:"studentName"`
		CategoryId         string    `json:"categoryId"`
		CreatorId          string    `json:"creatorId,omitempty"`
		CreatorName        string    `json:"creatorName,omitempty"`
		LongDesc           string    `json:"longDesc"`
		ShortDesc          string    `json:"shortDesc"`
		CreatedDate        time.Time `json:"createdDate"`
		EventTime          time.Time `json:"eventTime,omitempty"`
		Area               *area     `json:"area,omitempty"`
		Images             []image   `json:"images"`
		VisibleToGuardians bool      `json:"visibleToGuardians"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		queries := r.URL.Query()
		searchQuery := queries.Get("search")
		startDateQuery := queries.Get("startDate")
		endDateQuery := queries.Get("endDate")

		plan := sentry.StartSpan(r.Context(), "query_observations")
		observations, err := store.GetObservations(studentId, searchQuery, startDateQuery, endDateQuery)
		plan.Finish()

		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "Fail to query students",
				Error:   err,
			}
		}

		response := make([]responseBody, len(observations))
		for i, o := range observations {
			response[i].Id = o.Id
			response[i].StudentName = o.Student.Name
			response[i].CategoryId = o.CategoryId
			response[i].LongDesc = o.LongDesc
			response[i].ShortDesc = o.ShortDesc
			response[i].EventTime = o.EventTime
			response[i].CreatedDate = o.CreatedDate
			response[i].VisibleToGuardians = o.VisibleToGuardians
			if o.AreaId != uuid.Nil {
				response[i].Area = &area{
					Id:   o.Area.Id,
					Name: o.Area.Name,
				}
			}
			if o.CreatorId != "" {
				response[i].CreatorId = o.CreatorId
				response[i].CreatorName = o.Creator.Name
			}
			response[i].Images = make([]image, 0)
			for j := range o.Images {
				item := o.Images[j]
				response[i].Images = append(response[i].Images, image{
					Id:           item.Id,
					ThumbnailUrl: imgproxy.GenerateUrlFromS3(item.ObjectKey, 80, 80),
					OriginalUrl:  imgproxy.GenerateOriginalUrlFromS3(item.ObjectKey),
				})
			}
		}

		if err := rest.WriteJson(w, response); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func getMaterialProgress(s rest.Server, store Store) http.Handler {
	type responseBody struct {
		AreaId       string    `json:"areaId"`
		MaterialName string    `json:"materialName"`
		MaterialId   string    `json:"materialId"`
		Stage        int       `json:"stage"`
		UpdatedAt    time.Time `json:"updatedAt"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")
		//areaId := r.URL.Query().Get("areaId")

		progress, err := store.GetProgress(studentId)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "Failed querying material",
				Error:   err,
			}
		}

		// return empty array when there is no data
		response := make([]responseBody, 0)
		for _, progress := range progress {
			response = append(response, responseBody{
				AreaId:       progress.Material.Subject.Area.Id,
				MaterialName: progress.Material.Name,
				MaterialId:   progress.MaterialId,
				Stage:        progress.Stage,
				UpdatedAt:    progress.UpdatedAt,
			})
		}

		if err := rest.WriteJson(w, response); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func upsertMaterialProgress(s rest.Server, store Store) http.Handler {
	type responseBody struct {
		AreaId       string    `json:"areaId"`
		MaterialName string    `json:"materialName"`
		MaterialId   string    `json:"materialId"`
		Stage        int       `json:"stage"`
		UpdatedAt    time.Time `json:"updatedAt"`
	}
	type requestBody struct {
		Stage int `json:"stage"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")
		materialId := chi.URLParam(r, "materialId")

		var requestBody requestBody
		if err := rest.ParseJson(r.Body, &requestBody); err != nil {
			return rest.NewParseJsonError(err)
		}

		progress, err := store.UpdateProgress(postgres.StudentMaterialProgress{
			MaterialId: materialId,
			StudentId:  studentId,
			Stage:      requestBody.Stage,
			UpdatedAt:  time.Now(),
		})
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "Failed updating progress",
				Error:   err,
			}
		}

		if err := rest.WriteJson(w, &responseBody{
			AreaId:       progress.Material.Subject.Area.Id,
			MaterialName: progress.Material.Name,
			MaterialId:   progress.MaterialId,
			Stage:        progress.Stage,
			UpdatedAt:    progress.UpdatedAt,
		}); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func getPlans(s rest.Server, store Store) http.Handler {
	type Area struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type User struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type responseBody struct {
		Id          string    `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Date        time.Time `json:"date"`
		Area        *Area     `json:"area,omitempty"`
		User        User      `json:"user,omitempty"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")
		date := r.URL.Query().Get("date")

		parsedDate, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusBadRequest,
				Message: "date needs to be in ISO format",
				Error:   err,
			}
		}

		lessonPlans, err := store.GetLessonPlans(studentId, parsedDate)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "Failed to query lesson plan",
				Error:   err,
			}
		}

		response := make([]responseBody, len(lessonPlans))
		for i, plan := range lessonPlans {
			response[i] = responseBody{
				Id:          plan.Id,
				Title:       plan.LessonPlanDetails.Title,
				Description: plan.LessonPlanDetails.Description,
				Date:        *plan.Date,
				User: User{
					Id:   plan.LessonPlanDetails.UserId,
					Name: plan.LessonPlanDetails.User.Name,
				},
			}
			if plan.LessonPlanDetails.AreaId != "" {
				response[i].Area = &Area{
					Id:   plan.LessonPlanDetails.AreaId,
					Name: plan.LessonPlanDetails.Area.Name,
				}
			}
		}
		if err := rest.WriteJson(w, response); err != nil {
			return rest.NewWriteJsonError(err)
		}

		return nil
	})
}

func postNewImage(s rest.Server, store Store) http.Handler {
	type responseBody struct {
		Id           string    `json:"id"`
		OriginalUrl  string    `json:"originalUrl"`
		ThumbnailUrl string    `json:"thumbnailUrl"`
		CreatedAt    time.Time `json:"createdAt"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			return &rest.Error{
				Code:    http.StatusBadRequest,
				Message: "failed to parse payload",
				Error:   richErrors.Wrap(err, "failed to parse response body"),
			}
		}

		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			return &rest.Error{
				Code:    http.StatusBadRequest,
				Message: "invalid payload",
				Error:   richErrors.Wrap(err, "invalid payload"),
			}
		}

		image, err := store.CreateImage(studentId, file, fileHeader)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "Failed create file",
				Error:   err,
			}
		}

		body := &responseBody{
			Id:           image.Id.String(),
			OriginalUrl:  imgproxy.GenerateOriginalUrlFromS3(image.ObjectKey),
			ThumbnailUrl: imgproxy.GenerateUrlFromS3(image.ObjectKey, 400, 400),
			CreatedAt:    time.Time{},
		}

		w.WriteHeader(http.StatusCreated)
		if err := rest.WriteJson(w, &body); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func getStudentImages(s rest.Server, store Store) rest.Handler {
	type imageJson struct {
		Id           uuid.UUID `json:"id"`
		OriginalUrl  string    `json:"originalUrl"`
		ThumbnailUrl string    `json:"thumbnailUrl"`
		CreatedAt    time.Time `json:"createdAt"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		images, err := store.FindStudentImages(studentId)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed to query student images",
				Error:   err,
			}
		}

		response := make([]imageJson, 0)
		for _, image := range images {
			originalUrl := imgproxy.GenerateOriginalUrlFromS3(image.ObjectKey)
			thumbnailUrl := imgproxy.GenerateUrlFromS3(image.ObjectKey, 400, 400)
			response = append(response, imageJson{
				Id:           image.Id,
				CreatedAt:    image.CreatedAt,
				OriginalUrl:  originalUrl,
				ThumbnailUrl: thumbnailUrl,
			})
		}

		if err := rest.WriteJson(w, &response); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func getStudentVideos(s rest.Server, store Store) http.Handler {
	type video struct {
		Id                   string    `json:"id"`
		PlaybackUrl          string    `json:"playbackUrl"`
		ThumbnailUrl         string    `json:"thumbnailUrl"`
		OriginalThumbnailUrl string    `json:"originalThumbnailUrl"`
		Status               string    `json:"status"`
		CreatedAt            time.Time `json:"createdAt"`
	}
	type responseBody []video

	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		videos, err := store.FindStudentVideos(studentId)
		if err != nil {
			return rest.NewInternalServerError(err, "failed to query videos")
		}

		response := make(responseBody, 0)
		for _, v := range videos {
			response = append(response, video{
				Id:           v.Id.String(),
				PlaybackUrl:  v.PlaybackUrl,
				ThumbnailUrl: v.ThumbnailUrl + "?height=400&width=400&fit_mode=smartcrop",
				//ThumbnailUrl:         imgproxy.GenerateUrlFromHttp(v.ThumbnailUrl, 400, 400),
				OriginalThumbnailUrl: v.ThumbnailUrl,
				Status:               v.Status,
				CreatedAt:            v.CreatedAt,
			})
		}

		if err := rest.WriteJson(w, &response); err != nil {
			return rest.NewWriteJsonError(err)
		}
		return nil
	})
}

func exportMaterialProgressPdf(s rest.Server, store Store) rest.Handler {
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		progress, err := store.GetProgress(studentId)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed querying material",
				Error:   err,
			}
		}

		curriculum, err := store.FindCurriculum(studentId)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed querying material",
				Error:   err,
			}
		}

		pdf, err := ExportCurriculumPdf(curriculum, progress)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed to write pdf response",
				Error:   err,
			}
		}

		w.Header().Set("Content-Type", "application/pdf")
		if err := pdf.Write(w); err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed to write pdf response",
				Error:   err,
			}
		}
		return nil
	})
}

func exportMaterialProgressCsv(s rest.Server, store Store) rest.Handler {
	type responseBody struct {
		Areas       string `csv:"Areas"`
		Subjects    string `csv:"Subjects"`
		Materials   string `csv:"Materials"`
		Assessments string `csv:"Assessments"`
	}
	return s.NewHandler(func(w http.ResponseWriter, r *http.Request) *rest.Error {
		studentId := chi.URLParam(r, "studentId")

		progress, err := store.GetProgress(studentId)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed querying material",
				Error:   err,
			}
		}

		curriculum, err := store.FindCurriculum(studentId)
		if err != nil {
			return &rest.Error{
				Code:    http.StatusInternalServerError,
				Message: "failed querying material",
				Error:   err,
			}
		}

		body := make([]responseBody, 0)
		for _, area := range curriculum.Areas {
			for _, subject := range area.Subjects {
				for _, material := range subject.Materials {
					line := responseBody{
						Areas:     area.Name,
						Subjects:  subject.Name,
						Materials: material.Name,
					}
					for _, materialProgress := range progress {
						if materialProgress.MaterialId == material.Id {
							line.Assessments = domain.GetAssessmentName(materialProgress.Stage)
							break
						}
					}
					body = append(body, line)
				}
			}
		}

		if err := rest.WriteCsv(w, body); err != nil {
			return rest.NewWriteCsvError(err)
		}
		return nil
	})
}
