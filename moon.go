package main

// https://github.com/chooper/moontool-go

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Calculate the moon phase for a given date.
//
// So here's the deal, this has been ported from [moon.py][1] by Kevin Turner
// which in turn was ported from [moontool.c][2] by John Walker. I don't
// actually understand any of the math very well, so in this package's
// current state, it's a line-by-line port-of-a-port.
//
// [1]: http://bazaar.launchpad.net/~keturn/py-moon-phase/trunk/annotate/head:/moon.py
// [2]: http://www.fourmilab.ch/moontool/

type MoonTime time.Time

const (
	// 1980 January 0.0 in JDN (Julian Day Number)
	Epoch = 2444238.5
	// Ecliptic longitude of the Sun at epoch 1980.0
	EclipticLongitudeEpoch = 278.833540
	// Ecliptic longitude of the Sun at perigee
	EclipticLongitudePerigee = 282.596403
	// Eccentricity of Earth's orbit
	Eccentricity = 0.016718
	// Semi-major axis of Earth's orbit, in kilometers
	SunSmaxis = 1.49585e8
	// Sun's angular size, in degrees, at semi-major axis distance
	SunAngularSizeSmaxis = 0.533128
	// Moon's mean longitude at the epoch
	MoonMeanLongitudeEpoch = 64.975464
	// Mean longitude of the perigee at the epoch
	MoonMeanPerigeeEpoch = 349.383063
	// Synodic month (new Moon to new Moon), in days
	SynodicMonth = 29.53058868
)

var alphabet = [][]string{
	// 1 - NEW
	[]string{"🌝", "🌕"},
	// 6 - WAXING CRESCENT
	[]string{"🌖"},
	// 10 - FIRST QUARTER
	[]string{"🌗"},
	// 14 - WAXING GIBBOUS
	[]string{"🌘"},
	// 16 - FULL
	[]string{"🌚", "🌑"},
	// 21 - WANING GIBBOUS
	[]string{"🌒"},
	// 25 - LAST QUARTER
	[]string{"🌓"},
	// 29 - WANING CRESCENT
	[]string{"🌔"},
}

// fixangle
// I don't know what this actually does.
func fixangle(a float64) float64 {
	return a - 360.0*math.Floor(a/360.0)
}

// torad
// Convert a float from degrees to radians
func torad(d float64) float64 {
	return d * math.Pi / 180.0
}

// todeg
// Convert a float from radians to degrees
func todeg(r float64) float64 {
	return r * 180.0 / math.Pi
}

// Kepler's equation
// ???
func kepler(m, ecc float64) float64 {
	epsilon := 1e-6
	m = torad(m)
	e := m
	var delta float64
	for {
		delta = e - ecc*math.Sin(e) - m
		e = e - delta/(1.0-ecc*math.Cos(e))
		if math.Abs(delta) <= epsilon {
			break
		}
	}
	return e
}

// julianDayNumber returns the time's Julian Day Number
// relative to the epoch 12:00 January 1, 4713 BC, Monday.
// Stolen from: https://code.google.com/p/go/source/browse/src/pkg/time/time.go?name=weekly.2011-11-18#241
func julianDayNumber(year int64, month, day int) float64 {
	a := int64(14-month) / 12
	y := year + 4800 - a
	m := int64(month) + 12*a - 3
	return float64(int64(day) + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045)
}

// JulianDayNumber
// Get the JulianDayNumber for a given (Moon)Time
func (mt MoonTime) JulianDayNumber() float64 {
	t := time.Time(mt)
	return julianDayNumber(int64(t.Year()), int(t.Month()), t.Day())
}

func moon() string {
	phase := int(MoonTime(time.Now()).MoonAge() * float64(len(alphabet)))
	rand.Seed(time.Now().Unix())
	moons := alphabet[phase]
	return fmt.Sprintf("%%F{$MOON_COLOR}%s%%f", moons[rand.Intn(len(moons))])
}

// MoonAge
// Get the age of the moon for a given (Moon)Time
func (mt MoonTime) MoonAge() float64 {
	day := mt.JulianDayNumber() - Epoch
	mean_sun_anom := fixangle((360 / 365.2422) * day)
	// Convert from perigee coordinates to epoch 1980
	M := fixangle(mean_sun_anom + EclipticLongitudeEpoch - EclipticLongitudePerigee)
	// Kepler's equation?
	Ec := kepler(M, Eccentricity)
	Ec = math.Sqrt((1+Eccentricity)/(1-Eccentricity)) * math.Tan(Ec/2.0)
	// True anomaly
	Ec = 2 * todeg(math.Atan(Ec))

	// Suns's geometric ecliptic longuitude
	lambda_sun := fixangle(Ec + EclipticLongitudePerigee)

	// Calculation of moon's position
	moon_longitude := fixangle(13.1763966*day + MoonMeanLongitudeEpoch)
	mean_moon_anum := fixangle(moon_longitude - 0.1114041*day - MoonMeanPerigeeEpoch)

	evection := 1.2739 * math.Sin(torad(2*(moon_longitude-lambda_sun)-mean_moon_anum))

	// Annual equation
	annual_eq := 0.1858 * math.Sin(torad(M))

	// Correction term
	A3 := 0.37 * math.Sin(torad(M))
	MmP := mean_moon_anum + evection - annual_eq - A3

	// Correction for the equation of the centre
	mEc := 6.2886 * math.Sin(torad(MmP))
	// Another correction term
	A4 := 0.214 * math.Sin(torad(2*MmP))
	// Corrected longitude
	lP := moon_longitude + evection + mEc - annual_eq + A4
	// Variation
	variation := 0.6583 * math.Sin(torad(2*(lP-lambda_sun)))
	// True longitude
	lPP := lP + variation

	// Calculation of the phase of the Moon
	moon_age := lPP - lambda_sun
	age := fixangle(moon_age) / 360.0
	return age
}
