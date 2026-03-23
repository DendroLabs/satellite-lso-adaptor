"""
Telesat Lightspeed constellation model.

Models the 198-satellite LEO constellation at 1,325 km altitude
across 27 orbital planes with 78° inclination.

Public parameters sourced from Telesat's FCC filings and press releases.
"""

import math
from dataclasses import dataclass

EARTH_RADIUS_KM = 6371.0
CONSTELLATION_ALTITUDE_KM = 1325.0
NUM_SATELLITES = 198
NUM_PLANES = 27
SATS_PER_PLANE = NUM_SATELLITES // NUM_PLANES  # ~7 per plane (198/27 ≈ 7.3)
INCLINATION_DEG = 78.0
ISL_COUNT_PER_SAT = 4
ISL_BANDWIDTH_GBPS = 10.0

# Known landing stations from public announcements
LANDING_STATIONS = [
    {"id": "ls-gatineau", "name": "Gatineau TOC", "country": "Canada",
     "lat": 45.4765, "lon": -75.7013, "status": "operational"},
    {"id": "ls-france", "name": "France Landing Station", "country": "France",
     "lat": 48.8566, "lon": 2.3522, "status": "under-construction"},
    {"id": "ls-australia", "name": "Australia Landing Station", "country": "Australia",
     "lat": -33.8688, "lon": 151.2093, "status": "under-construction"},
    {"id": "ls-allan-park", "name": "Allan Park Teleport", "country": "Canada",
     "lat": 44.2167, "lon": -80.8167, "status": "operational"},
]


@dataclass
class LatLon:
    lat: float64
    lon: float64


def haversine_km(a: LatLon, b: LatLon) -> float:
    """Great-circle distance between two points in km."""
    dlat = math.radians(b.lat - a.lat)
    dlon = math.radians(b.lon - a.lon)
    x = (math.sin(dlat / 2) ** 2 +
         math.cos(math.radians(a.lat)) * math.cos(math.radians(b.lat)) *
         math.sin(dlon / 2) ** 2)
    c = 2 * math.atan2(math.sqrt(x), math.sqrt(1 - x))
    return EARTH_RADIUS_KM * c


def nearest_landing_station(point: LatLon) -> dict:
    """Find the nearest landing station to a geographic point."""
    best = None
    best_dist = float("inf")
    for ls in LANDING_STATIONS:
        d = haversine_km(point, LatLon(ls["lat"], ls["lon"]))
        if d < best_dist:
            best_dist = d
            best = ls
    return best


def estimate_satellite_hops(ground_dist_km: float) -> int:
    """Estimate ISL hops based on ground distance.

    Average inter-satellite spacing at 1325km with 198 sats / 27 planes
    is approximately 2000-3000km per hop along the ISL mesh.
    """
    avg_hop_km = 2500.0
    return max(1, math.ceil(ground_dist_km / avg_hop_km))


def is_within_coverage(point: LatLon) -> bool:
    """Check if a point falls within Lightspeed coverage.

    The constellation at 78° inclination provides coverage roughly
    from 78°S to 78°N latitude with some margin from elevation angles.
    Polar regions beyond ~75° may have limited service.
    """
    return abs(point.lat) <= 75.0
