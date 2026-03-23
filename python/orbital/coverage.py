"""
Coverage analysis for Telesat Lightspeed.

Determines if two geographic points can be connected via the Lightspeed
constellation and provides path details.
"""

from .constellation import (
    LatLon, haversine_km, nearest_landing_station,
    estimate_satellite_hops, is_within_coverage,
)
from .latency import estimate_one_way_latency


def check_coverage(lat_a: float, lon_a: float,
                   lat_z: float, lon_z: float) -> dict:
    """Check if two points can be connected via Telesat Lightspeed.

    Returns coverage feasibility, estimated latency, path details,
    and nearest landing stations.
    """
    a = LatLon(lat_a, lon_a)
    z = LatLon(lat_z, lon_z)

    a_covered = is_within_coverage(a)
    z_covered = is_within_coverage(z)
    feasible = a_covered and z_covered

    ground_dist = haversine_km(a, z)
    hops = estimate_satellite_hops(ground_dist)
    latency = estimate_one_way_latency(a, z, hops)

    nearest_a = nearest_landing_station(a)
    nearest_z = nearest_landing_station(z)

    # Build path description
    if feasible:
        path_desc = (
            f"Terminal A ({lat_a:.2f}, {lon_a:.2f}) -> "
            f"LEO satellite uplink -> "
            f"{hops} ISL hop(s) at 1325km altitude -> "
            f"LEO satellite downlink -> "
            f"Terminal Z ({lat_z:.2f}, {lon_z:.2f}). "
            f"Ground distance: {ground_dist:.0f}km, "
            f"estimated latency: {latency['totalMs']:.1f}ms"
        )
    else:
        reasons = []
        if not a_covered:
            reasons.append(f"Point A ({lat_a:.2f}, {lon_a:.2f}) outside coverage zone")
        if not z_covered:
            reasons.append(f"Point Z ({lat_z:.2f}, {lon_z:.2f}) outside coverage zone")
        path_desc = "Not feasible: " + "; ".join(reasons)

    return {
        "feasible": feasible,
        "estimatedLatencyMs": latency["totalMs"],
        "satelliteHops": hops,
        "nearestLandingA": nearest_a["id"] if nearest_a else "",
        "nearestLandingZ": nearest_z["id"] if nearest_z else "",
        "pathDescription": path_desc,
        "latencyBreakdown": latency,
        "groundDistanceKm": ground_dist,
    }
