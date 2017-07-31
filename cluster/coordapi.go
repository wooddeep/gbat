package cluster


type ICoordApi interface {
	Init(coordAddr string, coordPort int)
	Set (key, val string, para CoordPara) (interface{}, error)
	Get (key string, para CoordPara) (interface{}, error)
	GetDir(key string, para CoordPara) ([]string, error)
	Watch (key string, para CoordPara) (interface{})
	WatchHandler(w interface{}, para CoordPara, reg func(string, string, int, int))
}
