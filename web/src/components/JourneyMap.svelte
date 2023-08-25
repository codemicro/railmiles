<script>
    import {onMount} from "svelte";
    import L from "leaflet";
    import "leaflet-easybutton";

    export let geoJSON
    export let mapHeight = "480px"

    let map
    let geoJSONLayer

    const updateGeoJSON = (obj) => {
        if (map) {
            geoJSONLayer.removeFrom(map)
        }

        geoJSONLayer = L.geoJSON(obj, { onEachFeature: (feature, layer) => {
                if (feature.properties && feature.properties.name) {
                    layer.bindPopup(feature.properties.name);
                }
            }
        })

        if (map) {
            geoJSONLayer.addTo(map)
            centerMap()
        }
    }

    const centerMap = () => {
        if (map) {
            try {
                map.fitBounds(geoJSONLayer.getBounds()) // sometimes this errors, sometimes it doesn't
            } catch {}
        }
    }

    $: updateGeoJSON(geoJSON)

    onMount(() => {
        map = L.map("journey-map").setView([55.093, -2.894], 5);

        L.easyButton("bi-bullseye", centerMap).addTo(map);

        L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 19,
            attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        }).addTo(map);

        let ormOverlay = L.tileLayer('http://{s}.tiles.openrailwaymap.org/standard/{z}/{x}/{y}.png', {
            attribution: '<a href="https://www.openstreetmap.org/copyright">© OpenStreetMap contributors</a>, Style: <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA 2.0</a> <a href="http://www.openrailwaymap.org/">OpenRailwayMap</a> and OpenStreetMap',
            minZoom: 2,
            maxZoom: 19,
            tileSize: 256,
            className: "tile-orm",
        });

        L.control.layers({}, {"OpenRailwayMap": ormOverlay}).addTo(map);

        updateGeoJSON(geoJSON)
    })
</script>

<div id="journey-map" style="--journey-map-height: {mapHeight}"></div>

<style>
    #journey-map {
        height: var(--journey-map-height);
    }

    :global(#journey-map .tile-orm .leaflet-tile) {
        filter: grayscale(1);
    }
</style>