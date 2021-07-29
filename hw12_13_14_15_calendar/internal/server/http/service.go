package internalhttp

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/gorilla/mux"
)

var ErrUnsupportedMediaType = errors.New("found unsupported media type, application/json expected")

type Service struct {
	app server.Application
}

type CreateEventData struct {
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
}

func (s Service) AddEventHandler(w http.ResponseWriter, r *http.Request) {
	eventData := new(CreateEventData)
	if err := receiveJSON(r, eventData); err != nil {
		if errors.Is(err, ErrUnsupportedMediaType) {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event, err := s.app.CreateEvent(
		r.Context(),
		eventData.Title,
		eventData.StartTime,
		eventData.EndTime,
		eventData.Description,
		eventData.OwnerID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := sendJSON(w, &event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s Service) UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	event := new(storage.Event)
	if err := receiveJSON(r, event); err != nil {
		if errors.Is(err, ErrUnsupportedMediaType) {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := s.app.UpdateEvent(r.Context(), *event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := sendJSON(w, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s Service) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	eventID, ok := mux.Vars(r)["eventId"]
	if !ok || eventID == "" {
		http.Error(w, "eventID route param is required", http.StatusBadRequest)
		return
	}

	err := s.app.DeleteEvent(r.Context(), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Service) FindEventsHandler(w http.ResponseWriter, r *http.Request) {
	routeParams := mux.Vars(r)

	year, _ := strconv.Atoi(routeParams["year"])
	month, _ := strconv.Atoi(routeParams["month"])
	day, _ := strconv.Atoi(routeParams["day"])
	requestedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	var events []storage.Event
	var err error
	switch periodStr := routeParams["period"]; periodStr {
	case "day":
		events, err = s.app.ListDayEvents(r.Context(), requestedDate)
	case "week":
		events, err = s.app.ListWeekEvents(r.Context(), requestedDate)
	case "month":
		events, err = s.app.ListMonthEvents(r.Context(), requestedDate)
	default:
		http.Error(w, "not valid period param", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := sendJSON(w, events); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// receiveJSON reads JSON request into v.
// C'mon golang why i need manually do this for all my http handlers? (More important TEST IT all the time >_<)
// Maybe it's fun to do this in every project (and TEST IT in every project).
func receiveJSON(r *http.Request, v interface{}) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	if mediaType != "application/json" {
		return ErrUnsupportedMediaType
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

// sendJSON writes JSON response into w.
// C'mon golang why i need manually do this for all my http handlers? (More important TEST IT all the time >_<)
// Maybe it's fun to do this in every project (and TEST IT in every project).
func sendJSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		return err
	}
	return nil
}
