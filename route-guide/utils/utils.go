package utils

import (
	"fmt"
	"math"

	"wesionary.team/dipeshdulal/route-guide/mrouteguide"
)

// ToRadians converts number to radians
func ToRadians(num float64) float64 {
	return num * math.Pi / float64(180)
}

// CalcDistance calculates distance between two points
func CalcDistance(p1 *mrouteguide.Point, p2 *mrouteguide.Point) int32 {
	const CordFactor float64 = 1e7
	const R = float64(63710)
	lat1 := ToRadians(float64(p1.Latitude) / CordFactor)
	lat2 := ToRadians(float64(p2.Latitude) / CordFactor)
	lng1 := ToRadians(float64(p2.Longitude) / CordFactor)
	lng2 := ToRadians(float64(p2.Longitude) / CordFactor)
	dlat := lat2 - lat1
	dlng := lng2 - lng1

	// Haversine formula for calculating distance between long and lat coord
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return int32(distance)
}

// InRange finds if point in range
// long -> x, lat -> y
func InRange(point *mrouteguide.Point, rect *mrouteguide.Rectangle) bool {

	// bounding planes
	left := math.Min(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	right := math.Max(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	top := math.Max(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
	bottom := math.Min(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))

	if float64(point.Longitude) >= left &&
		float64(point.Longitude) <= right &&
		float64(point.Latitude) <= top &&
		float64(point.Latitude) >= bottom {
		return true
	}

	return false

}

// Serialize serializes
func Serialize(point *mrouteguide.Point) string {
	return fmt.Sprintf("%d %d", point.Latitude, point.Longitude)
}
