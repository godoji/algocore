package algo

import "github.com/godoji/algocore/pkg/env"

type StepFunction = func(chart env.MarketSupplier, term *ResultHandler, mem *env.Memory, params env.Parameters)

type ResultSet struct {
	Symbols map[string]*SymbolResultSet `json:"symbols"`
}

type SymbolResultSet struct {
	Scenarios []*ScenarioResultSet `json:"scenarios"`
}

type ScenarioResultSet struct {
	Events     []*ResultEvent `json:"events"`
	Parameters []float64      `json:"parameters"`
}

type PointAnnotation struct {
	Time  int64   `json:"time"`
	Price float64 `json:"price"`
	Icon  string  `json:"icon"`
	Color string  `json:"color"`
}

type LineAnnotation struct {
	TimeBegin  int64   `json:"timeFrom"`
	TimeEnd    int64   `json:"timeEnd"`
	PriceBegin float64 `json:"priceBegin"`
	PriceEnd   float64 `json:"priceEnd"`
	Style      string  `json:"style"`
	Color      string  `json:"color"`
}

type LabelAnnotation struct {
	Text  string  `json:"text"`
	Time  int64   `json:"time"`
	Price float64 `json:"price"`
	Color string  `json:"color"`
}

type AnnotationCollection struct {
	Points []*PointAnnotation `json:"points"`
	Lines  []*LineAnnotation  `json:"lines"`
	Labels []*LabelAnnotation `json:"labels"`
}

type ResultEvent struct {
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
	results   *ScenarioResultSet
}

type ResultEventHandler struct {
	event *ResultEvent
}

func (r *ResultEventHandler) AddLine(line *LineAnnotation) *ResultEventHandler {
	if r.event.Annotations == nil {
		r.event.Annotations = NewAnnotationCollection()
	}
	r.event.Annotations.Lines = append(r.event.Annotations.Lines, line)
	return r
}

func (r *ResultEventHandler) AddPoint(point *PointAnnotation) *ResultEventHandler {
	if r.event.Annotations == nil {
		r.event.Annotations = NewAnnotationCollection()
	}
	r.event.Annotations.Points = append(r.event.Annotations.Points, point)
	return r
}

func (r *ResultEventHandler) AddLabel(label *LabelAnnotation) *ResultEventHandler {
	if r.event.Annotations == nil {
		r.event.Annotations = NewAnnotationCollection()
	}
	r.event.Annotations.Labels = append(r.event.Annotations.Labels, label)
	return r
}

func (r *ResultEventHandler) SetColor(color string) *ResultEventHandler {
	r.event.Color = color
	return r
}

func (r *ResultEventHandler) SetIcon(icon string) *ResultEventHandler {
	r.event.Icon = icon
	return r
}

func (r *ResultEventHandler) SetPrice(price float64) *ResultEventHandler {
	r.event.Price = price
	return r
}

func (r *ResultEventHandler) SetTime(ts int64) *ResultEventHandler {
	r.event.Time = ts
	return r
}

func NewResultHandler(res *ScenarioResultSet, ts int64, price float64) *ResultHandler {
	return &ResultHandler{results: res, timestamp: ts, price: price}
}

func NewAnnotationCollection() *AnnotationCollection {
	return &AnnotationCollection{
		Points: make([]*PointAnnotation, 0),
		Lines:  make([]*LineAnnotation, 0),
		Labels: make([]*LabelAnnotation, 0),
	}
}

func (r *ResultHandler) NewEvent(label string) *ResultEventHandler {
	e := &ResultEvent{
		CreatedOn:   r.timestamp,
		Time:        r.timestamp,
		Price:       r.price,
		Label:       label,
		Icon:        "event",
		Color:       "blue",
		Annotations: nil,
	}
	r.results.Events = append(r.results.Events, e)
	return &ResultEventHandler{event: e}
}
