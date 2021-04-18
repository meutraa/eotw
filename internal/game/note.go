package game

type Note struct {
	Index   int
	Row     int64
	Denom   int
	Hit     bool
	IsMine  bool
	Miss    bool
	HitTime int64
	Ms      int64
}
