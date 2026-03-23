"""Tests for the latency estimation engine."""
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from orbital.latency import (
    estimate_one_way_latency, slant_range_km,
    PT5_MAX_ONE_WAY_DELAY_MS,
    CONSTELLATION_ALTITUDE_KM, MIN_ELEVATION_DEG,
)
from orbital.constellation import LatLon


class TestSlantRange:
    def test_positive(self):
        slant = slant_range_km(CONSTELLATION_ALTITUDE_KM, MIN_ELEVATION_DEG)
        assert slant > 0

    def test_decreases_with_elevation(self):
        low = slant_range_km(CONSTELLATION_ALTITUDE_KM, 20.0)
        high = slant_range_km(CONSTELLATION_ALTITUDE_KM, 80.0)
        assert high < low, "higher elevation should have shorter slant range"

    def test_90_degrees_equals_altitude(self):
        slant = slant_range_km(CONSTELLATION_ALTITUDE_KM, 90.0)
        assert abs(slant - CONSTELLATION_ALTITUDE_KM) < 1.0


class TestEstimateOneWayLatency:
    def test_ottawa_paris(self):
        a = LatLon(45.4215, -75.6972)
        z = LatLon(48.8566, 2.3522)
        result = estimate_one_way_latency(a, z)
        total = result["totalMs"]
        assert 15 < total < 80, f"Ottawa-Paris should be ~25-40ms, got {total:.1f}ms"

    def test_returns_all_components(self):
        a = LatLon(45.0, -75.0)
        z = LatLon(48.0, 2.0)
        result = estimate_one_way_latency(a, z)
        assert "totalMs" in result
        assert "uplinkDownlinkMs" in result
        assert "atmosphericMs" in result
        assert "islMs" in result
        assert "processingMs" in result
        assert "groundDistanceKm" in result
        assert "satelliteHops" in result
        assert "slantRangeKm" in result
        assert "pt5Valid" in result

    def test_all_components_positive(self):
        a = LatLon(45.0, -75.0)
        z = LatLon(48.0, 2.0)
        result = estimate_one_way_latency(a, z)
        assert result["uplinkDownlinkMs"] > 0
        assert result["atmosphericMs"] > 0
        assert result["islMs"] > 0
        assert result["processingMs"] > 0

    def test_components_sum_to_total(self):
        a = LatLon(45.0, -75.0)
        z = LatLon(48.0, 2.0)
        result = estimate_one_way_latency(a, z)
        component_sum = (result["uplinkDownlinkMs"] + result["atmosphericMs"] +
                         result["islMs"] + result["processingMs"])
        assert abs(component_sum - result["totalMs"]) < 0.1

    def test_more_hops_more_latency(self):
        a = LatLon(45.0, -75.0)
        z = LatLon(50.0, -70.0)
        one_hop = estimate_one_way_latency(a, z, hops=1)
        three_hop = estimate_one_way_latency(a, z, hops=3)
        assert three_hop["totalMs"] > one_hop["totalMs"]

    def test_short_path_pt5_valid(self):
        a = LatLon(45.0, -75.0)
        z = LatLon(46.0, -74.0)
        result = estimate_one_way_latency(a, z)
        assert result["pt5Valid"], f"short path should be PT5 valid, got {result['totalMs']:.1f}ms"

    def test_custom_hops(self):
        a = LatLon(45.0, -75.0)
        z = LatLon(48.0, 2.0)
        result = estimate_one_way_latency(a, z, hops=5)
        assert result["satelliteHops"] == 5

    def test_same_location(self):
        a = LatLon(45.0, -75.0)
        result = estimate_one_way_latency(a, a, hops=1)
        # Should still have uplink/downlink delay
        assert result["totalMs"] > 0
        assert result["uplinkDownlinkMs"] > 0
        # ISL should be minimal for same location
        assert result["islMs"] < 1.0
