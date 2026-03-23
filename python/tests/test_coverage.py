"""Tests for the coverage analysis module."""
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from orbital.coverage import check_coverage


class TestCheckCoverage:
    def test_ottawa_paris_feasible(self):
        result = check_coverage(45.4215, -75.6972, 48.8566, 2.3522)
        assert result["feasible"]
        assert result["estimatedLatencyMs"] > 0
        assert result["satelliteHops"] >= 1
        assert result["nearestLandingA"] == "ls-gatineau"
        assert result["nearestLandingZ"] == "ls-france"

    def test_ottawa_yellowknife_feasible(self):
        result = check_coverage(45.4215, -75.6972, 62.454, -114.3718)
        assert result["feasible"]
        assert result["nearestLandingA"] == "ls-gatineau"

    def test_polar_infeasible(self):
        result = check_coverage(85.0, 0.0, 45.0, -75.0)
        assert not result["feasible"]
        assert "outside coverage zone" in result["pathDescription"]

    def test_both_polar_infeasible(self):
        result = check_coverage(85.0, 0.0, -85.0, 0.0)
        assert not result["feasible"]

    def test_response_structure(self):
        result = check_coverage(45.0, -75.0, 48.0, 2.0)
        assert "feasible" in result
        assert "estimatedLatencyMs" in result
        assert "satelliteHops" in result
        assert "nearestLandingA" in result
        assert "nearestLandingZ" in result
        assert "pathDescription" in result
        assert "latencyBreakdown" in result
        assert "groundDistanceKm" in result

    def test_path_description_feasible(self):
        result = check_coverage(45.0, -75.0, 48.0, 2.0)
        assert "ISL hop" in result["pathDescription"]
        assert "1325km" in result["pathDescription"]

    def test_ground_distance_realistic(self):
        result = check_coverage(45.4215, -75.6972, 48.8566, 2.3522)
        assert 5600 < result["groundDistanceKm"] < 5700

    def test_latency_breakdown_present(self):
        result = check_coverage(45.0, -75.0, 48.0, 2.0)
        breakdown = result["latencyBreakdown"]
        assert "totalMs" in breakdown
        assert "uplinkDownlinkMs" in breakdown
        assert "islMs" in breakdown
