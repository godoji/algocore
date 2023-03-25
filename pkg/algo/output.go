package algo

type ResultSet struct {
	Symbols map[string]*SymbolResultSet `json:"symbols"`
}

type SymbolResultSet struct {
	Scenarios []*ScenarioSet `json:"scenarios"`
}

type ScenarioSet struct {
	Events     []*Event  `json:"events"`
	Parameters []float64 `json:"parameters"`
}

type PointAnnotation struct {
	Text  string  `json:"text"`
	Time  int64   `json:"time"`
	Price float64 `json:"price"`
	Icon  string  `json:"icon"`
	Color string  `json:"color"`
}

type SegmentAnnotation struct {
	TimeBegin  int64   `json:"timeFrom"`
	TimeEnd    int64   `json:"timeEnd"`
	PriceBegin float64 `json:"priceBegin"`
	PriceEnd   float64 `json:"priceEnd"`
	Style      string  `json:"style"`
	Color      string  `json:"color"`
}

type AnnotationCollection struct {
	Points   []*PointAnnotation   `json:"points"`
	Segments []*SegmentAnnotation `json:"segments"`
}

type Event struct {
	CreatedOn   int64                 `json:"createdOn"`
	Time        int64                 `json:"time"`
	Price       float64               `json:"price"`
	Label       string                `json:"label"`
	Icon        string                `json:"icon"`
	Color       string                `json:"color"`
	Annotations *AnnotationCollection `json:"annotations"`
}

type ResultHandler struct {
	timestamp int64
	price     float64
	results   *ScenarioSet
}

type EventHandler struct {
	event *Event
}

func (r EventHandler) AddSegment(segment *SegmentAnnotation) EventHandler {
	if r.event.Annotations == nil {
		r.event.Annotations = NewAnnotationCollection()
	}
	r.event.Annotations.Segments = append(r.event.Annotations.Segments, segment)
	return r
}

func (r EventHandler) AddPoint(point *PointAnnotation) EventHandler {
	if r.event.Annotations == nil {
		r.event.Annotations = NewAnnotationCollection()
	}
	r.event.Annotations.Points = append(r.event.Annotations.Points, point)
	return r
}

func (r EventHandler) SetColor(color string) EventHandler {
	r.event.Color = color
	return r
}

func (r EventHandler) SetIcon(icon string) EventHandler {
	r.event.Icon = icon
	return r
}

func (r EventHandler) SetPrice(price float64) EventHandler {
	r.event.Price = price
	return r
}

func (r EventHandler) SetTime(ts int64) EventHandler {
	r.event.Time = ts
	return r
}

func NewResultHandler(res *ScenarioSet, ts int64, price float64) *ResultHandler {
	return &ResultHandler{results: res, timestamp: ts, price: price}
}

func NewAnnotationCollection() *AnnotationCollection {
	return &AnnotationCollection{
		Points:   make([]*PointAnnotation, 0),
		Segments: make([]*SegmentAnnotation, 0),
	}
}

func (r *ResultHandler) NewEvent(label string) EventHandler {
	e := &Event{
		CreatedOn:   r.timestamp,
		Time:        r.timestamp,
		Price:       r.price,
		Label:       label,
		Icon:        "event",
		Color:       "white",
		Annotations: nil,
	}
	r.results.Events = append(r.results.Events, e)
	return EventHandler{event: e}
}
