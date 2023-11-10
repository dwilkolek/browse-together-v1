package internal

type SessionMemberDTO struct {
	GivenIdentifier string  `json:"givenIdentifier"`
	MemberId        int64   `json:"memberId"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	Selector        string  `json:"selector"`
	Location        string  `json:"location"`
}

type SessionDTO struct {
	Id                string `json:"id"`
	JoinUrl           string `json:"joinUrl"`
	Name              string `json:"name"`
	BaseUrl           string `json:"baseUrl"`
	CreatorIdentifier string `json:"creatorIdentifier"`
}

type UpdatePositionCmdDTO struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Selector string  `json:"selector"`
	Location string  `json:"location"`
}

type PositionStateDTO struct {
	MemberId int64   `json:"memberId"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Selector string  `json:"selector"`
	Location string  `json:"location"`
}
