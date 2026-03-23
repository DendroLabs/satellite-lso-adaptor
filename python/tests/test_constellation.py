"""Tests for the constellation model."""
import math
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from orbital.constellation import (
    LatLon, haversine_km, nearest_landing_station,
    estimate_satellite_hops, is_within_coverage,
    NUM_SATELLITES, NUM_PLANES, CONSTELLATION_ALTITUDE_KM,
    INCLINATION_DEG, LANDING_STATIONS,
)


class TestHaversineKm:
    def test_ottawa_paris(self):
        a = LatLon(45.4215, -75.6972)
        z = LatLon(48.8566, 2.3522)
        dist = haversine_km(a, z)
        assert 5600 < dist < 5700, f"Ottawa-Paris should be ~5650km, got {dist:.0f}"

    def test_ottawa_yellowknife(self):
        a = LatLon(45.4215, -75.6972)
        z = LatLon(62.454, -114.3718)
        dist = haversine_km(a, z)
        assert 2800 < dist < 3200, f"Ottawa-Yellowknife should be ~3000km, got {dist:.0f}"

    def test_same_point(self):
        a = LatLon(45.0, -75.0)
        assert haversine_km(a, a) < 0.001

    def test_antipodal(self):
        a = LatLon(0.0, 0.0)
        z = LatLon(0.0, 180.0)
        dist = haversine_km(a, z)
        assert 20000 < dist < 20100, f"antipodal should be ~20015km, got {dist:.0f}"

    def test_symmetry(self):
        a = LatLon(45.0, -75.0)
        z = LatLon(48.0, 2.0)
        assert abs(haversine_km(a, z) - haversine_km(z, a)) < 0.001


class TestNearestLandingStation:
    def test_ottawa_nearest_gatineau(self):
        ottawa = LatLon(45.4215, -75.6972)
        nearest = nearest_landing_station(ottawa)
        assert nearest["id"] == "ls-gatineau"

    def test_paris_nearest_france(self):
        paris = LatLon(48.8566, 2.3522)
        nearest = nearest_landing_station(paris)
        assert nearest["id"] == "ls-france"

    def test_sydney_nearest_australia(self):
        sydney = LatLon(-33.8688, 151.2093)
        nearest = nearest_landing_station(sydney)
        assert nearest["id"] == "ls-australia"

    def test_returns_dict(self):
        result = nearest_landing_station(LatLon(0, 0))
        assert "id" in result
        assert "lat" in result
        assert "lon" in result


class TestEstimateSatelliteHops:
    def test_short_distance(self):
        assert estimate_satellite_hops(500) == 1

    def test_medium_distance(self):
        assert estimate_satellite_hops(5000) == 2

    def test_long_distance(self):
        assert estimate_satellite_hops(10000) == 4

    def test_minimum_one_hop(self):
        assert estimate_satellite_hops(0) >= 1
        assert estimate_satellite_hops(100) >= 1


class TestIsWithinCoverage:
    def test_ottawa_covered(self):
        assert is_within_coverage(LatLon(45.42, -75.70))

    def test_paris_covered(self):
        assert is_within_coverage(LatLon(48.86, 2.35))

    def test_equator_covered(self):
        assert is_within_coverage(LatLon(0.0, 0.0))

    def test_arctic_edge(self):
        assert is_within_coverage(LatLon(74.0, 0.0))

    def test_north_pole_not_covered(self):
        assert not is_within_coverage(LatLon(85.0, 0.0))

    def test_south_pole_not_covered(self):
        assert not is_within_coverage(LatLon(-85.0, 0.0))

    def test_boundary(self):
        assert is_within_coverage(LatLon(75.0, 0.0))
        assert not is_within_coverage(LatLon(75.1, 0.0))


class TestConstellationConstants:
    def test_satellite_count(self):
        assert NUM_SATELLITES == 198

    def test_plane_count(self):
        assert NUM_PLANES == 27

    def test_altitude(self):
        assert CONSTELLATION_ALTITUDE_KM == 1325.0

    def test_inclination(self):
        assert INCLINATION_DEG == 78.0

    def test_landing_stations_count(self):
        assert len(LANDING_STATIONS) == 4

    def test_landing_station_structure(self):
        for ls in LANDING_STATIONS:
            assert "id" in ls
            assert "lat" in ls
            assert "lon" in ls
            assert "country" in ls
            assert "status" in ls
            assert -90 <= ls["lat"] <= 90
            assert -180 <= ls["lon"] <= 180
