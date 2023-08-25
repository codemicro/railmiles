<script>
    import BaseLayout from "../components/BaseLayout.svelte";
    import L from "leaflet";
    import { onMount } from "svelte";
    import Loading from "../components/Loading.svelte";
    import {makeURL, roundFloat} from "../util.js";
    import JourneyTable from "../components/JourneyTable.svelte";
    import JourneyMap from "../components/JourneyMap.svelte";

    let map;
    let stats = {
        lastMonth: {count: 0, miles: 0},
        ytd: {count: 0, miles: 0},
        allTime: {count: 0, miles: 0},
    };
    let journeys;
    let journeyGeoData;
    let ready = false;

    onMount(async () => {
        let response;
        try {
            response = await fetch(makeURL("/api/dashboard"));
        } catch (e) {
            alert(e.toString())
            return
        }

        const responseJSON = await response.json();

        stats = responseJSON.stats;
        journeys = responseJSON.journeys;

        journeyGeoData = responseJSON.geoJSON

        ready = true;
    })
</script>

<BaseLayout>
    {#if !ready}
        <Loading />
    {/if}

    <h1 class="pb-4"><i class="bi-speedometer2"></i> Dashboard</h1>

    <div class="row gap-2 g-2">
        <div class="col-sm card text-bg-primary">
            <div class="card-header">Last Month</div>
            <div class="card-body">
                <div class="d-flex text-center justify-content-center">
                    <div>
                        <span class="fs-2">{roundFloat(stats.lastMonth.miles, 1)}</span>
                        <span class="fs-5">miles</span>
                    </div>
                    <div>
                        <span class="fs-2">{stats.lastMonth.count}</span>
                        <span class="fs-5">journeys</span>
                    </div>
                </div>
            </div>
        </div>
        <div class="col-sm card text-bg-light">
            <div class="card-header">Year-to-Date</div>
            <div class="card-body">
                <div class="d-flex text-center justify-content-center">
                    <div>
                        <span class="fs-2">{roundFloat(stats.ytd.miles, 1)}</span>
                        <span class="fs-5">miles</span>
                    </div>
                    <div>
                        <span class="fs-2">{stats.ytd.count}</span>
                        <span class="fs-5">journeys</span>
                    </div>
                </div>
            </div>
        </div>
        <div class="col-sm card text-bg-light">
            <div class="card-header">All Time</div>
            <div class="card-body">
                <div class="d-flex text-center justify-content-center">
                    <div>
                        <span class="fs-2">{roundFloat(stats.allTime.miles, 1)}</span>
                        <span class="fs-5">miles</span>
                    </div>
                    <div>
                        <span class="fs-2">{stats.allTime.count}</span>
                        <span class="fs-5">journeys</span>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <h3 class="pt-4">Recent journeys</h3>

    <JourneyMap geoJSON={journeyGeoData} />

    <div class="pt-4"></div>

    <JourneyTable showMore=true journeys={journeys} />
</BaseLayout>

<style>
    .d-flex {
        gap: 1em;
    }
</style>