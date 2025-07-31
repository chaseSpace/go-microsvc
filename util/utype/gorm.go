package utype

import (
	"context"
	"fmt"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"regexp"
	"strings"

	"github.com/spf13/cast"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Geometry struct {
	*commonpb.Geometry
}

func (g *Geometry) GormDataType() string {
	return "geometry"
}

// GormValue for write
func (g *Geometry) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if g.Geometry == nil || g.Lat == "" || g.Lng == "" {
		return clause.Expr{SQL: "NULL"}
	}
	var lng, lat = strings.TrimSpace(g.Lng), strings.TrimSpace(g.Lat)
	return clause.Expr{
		SQL:  "ST_PointFromText(?)",
		Vars: []interface{}{fmt.Sprintf("POINT(%v %v)", lng, lat)},
	}
}

var pointRegex = regexp.MustCompile(`POINT\((-?\d+(\.\d+)?) (-?\d+(\.\d+)?)\)`)

// Scan for read
func (loc *Geometry) Scan(v interface{}) error {
	var pointStr string
	switch value := v.(type) {
	case []byte:
		pointStr = string(value)
	case string:
		pointStr = value
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	if len(pointStr) > 0 {
		ss := pointRegex.FindStringSubmatch(pointStr)
		//fmt.Printf("%+v", ss)
		if len(ss) == 5 {
			loc.Geometry = &commonpb.Geometry{
				Lng: cast.ToString(ss[1]),
				Lat: cast.ToString(ss[3]),
			}
		} else {
			return fmt.Errorf("invalid location point: (%s)", pointStr)
		}
	}
	return nil
}

var singlePointRegex = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

func (loc *Geometry) Check() error {
	if loc.Geometry != nil {
		loc.Lng, loc.Lat = strings.TrimSpace(loc.Lng), strings.TrimSpace(loc.Lat)
		s := loc.Lat + loc.Lng
		if len(s) == 0 {
			return nil
		}
		if len(s) > 0 && (len(s) == len(loc.Lat) || len(s) == len(loc.Lng)) {
			return xerr.ErrParams.New("field `geometry` contain a non-empty value and a empty value")
		}
		if !singlePointRegex.MatchString(loc.Lng) {
			return xerr.ErrParams.New("invalid location lng : (%s)", loc.Lng)
		}
		if !singlePointRegex.MatchString(loc.Lat) {
			return xerr.ErrParams.New("invalid location lat: (%s)", loc.Lat)
		}
	}
	return nil
}
