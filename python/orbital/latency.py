"""
Latency estimation for Telesat Lightspeed paths.

Models one-way latency based on:
- Uplink/downlink slant range at minimum elevation angle
- Inter-satellite link (ISL) propagation via optical links
- Processing delay per satellite hop
- Atmospheric propagation adjustment per MEF 23.2.2
"""

import math
from .constellation import (
    CONSTELLATION_ALTITUDE_KM, EARTH_RADIUS_KM,
    haversine_km, estimate_satellite_hops, LatLon,
)

SPEED_OF_LIGHT_KM_MS = 299.792
ATMOSPHERIC_ADJUSTMENT = 0.164  # 16.4% per MEF 23.2.2
ATMOSPHERIC_DELAY_MS_PER_KM = 0.0033
ISL_SPEED_FACTOR = 0.97  # optical ISL slightly slower than vacuum c
PROCESSING_DELAY_PER_HOP_MS = 0.5
MIN_ELEVATION_DEG = 40.0  # typical for Lightspeed

# MEF 23.2.2 PT5 limits
PT5_MAX_ONE_WAY_DELAY_MS = 150.0
PT5_MAX_JITTER_MS = 30.0
PT5_MAX_FRAME_LOSS_PCT = 0.1


def slant_range_km(altitude_km: float, elevation_deg: float) -> float:
    """Distance from ground to satellite at given elevation angle."""
    elev_rad = math.radians(elevation_deg)
    r = EARTH_RADIUS_KM
    h = altitude_km
    return (-r * math.sin(elev_rad) +
            math.sqrt(r**2 * math.sin(elev_rad)**2 + 2 * r * h + h**2))


def estimate_one_way_latency(a: LatLon, z: LatLon, hops: int | None = None) -> dict:
    """Estimate one-way latency for a Lightspeed path.

    Returns a dict with the breakdown of latency components.
    """
    ground_dist = haversine_km(a, z)

    if hops is None:
        hops = estimate_satellite_hops(ground_dist)

    slant = slant_range_km(CONSTELLATION_ALTITUDE_KM, MIN_ELEVATION_DEG)

    # Uplink + downlink
    updown_ms = (2 * slant) / SPEED_OF_LIGHT_KM_MS

    # Atmospheric delay
    atmo_ms = 2 * slant * ATMOSPHERIC_DELAY_MS_PER_KM * ATMOSPHERIC_ADJUSTMENT

    # ISL propagation
    isl_dist_km = (hops * (ground_dist / max(hops, 1)) *
                   (1 + CONSTELLATION_ALTITUDE_KM / EARTH_RADIUS_KM))
    isl_ms = isl_dist_km / (SPEED_OF_LIGHT_KM_MS * ISL_SPEED_FACTOR)

    # Processing
    proc_ms = hops * PROCESSING_DELAY_PER_HOP_MS

    total_ms = updown_ms + atmo_ms + isl_ms + proc_ms

    return {
        "totalMs": round(total_ms, 2),
        "uplinkDownlinkMs": round(updown_ms, 2),
        "atmosphericMs": round(atmo_ms, 2),
        "islMs": round(isl_ms, 2),
        "processingMs": round(proc_ms, 2),
        "groundDistanceKm": round(ground_dist, 1),
        "satelliteHops": hops,
        "slantRangeKm": round(slant, 1),
        "pt5Valid": total_ms <= PT5_MAX_ONE_WAY_DELAY_MS,
    }
