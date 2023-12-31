<script>
    import BaseLayout from "../components/BaseLayout.svelte";
    import {onMount} from "svelte";
    import {formatDate, makeURL, roundFloat} from "../util.js";
    import Loading from "../components/Loading.svelte";
    import JourneyMap from "../components/JourneyMap.svelte";
    import {push} from "svelte-spa-router";

    let ready = false
    let transparentLoading = false

    export let params = {
        id: undefined,
    }
    let journey;
    let geoJSON;

    const initialLoad = async () => {
        ready = false;
        let response;
        try {
            response = await fetch(makeURL("/api/journeys/" + params.id));
        } catch (e) {
            alert(e.toString())
            return
        }

        if (!response.ok) {
            await push("/notfound")
        }

        const responseJSON = await response.json()
        journey = responseJSON.data
        geoJSON = responseJSON.geoJSON
        ready = true
    }

    onMount(initialLoad)

    const switchToJourney = async (newID) => {
        transparentLoading = true
        params.id = newID
        await initialLoad()
    }

    const deleteSelf = async () => {
        if (!confirm("Are you sure you want to permanently delete this journey?")) {
            return
        }

        transparentLoading = true
        ready = false

        let response;
        try {
            response = await fetch(makeURL("/api/journeys/" + params.id), {method: "DELETE"});
        } catch (e) {
            alert(e.toString())
            return
        }

        if (!response.ok) {
            alert(response.statusText)
            return
        }

        alert("Success!")
        await push("/journeys")
    }

    const createReturn = async () => {
        transparentLoading = true
        ready = false

        let response;
        try {
            response = await fetch(makeURL("/api/journeys/" + params.id + "/return"), {method: "POST"});
        } catch (e) {
            alert(e.toString())
            return
        }

        if (!response.ok) {
            alert(response.statusText)
            return
        }

        const j = await response.json()

        alert("Success!")
        await switchToJourney(j)
    }
</script>

<BaseLayout>
    {#if !ready}
        <Loading transparent={transparentLoading}/>
    {/if}

    {#if journey}
        <h1 class="pb-4"><i class="bi-ticket-detailed"></i> {journey.from.full} to {journey.to.full}</h1>

        <JourneyMap geoJSON={geoJSON}/>

        <table class="table mt-4 mb-4">
            <tbody>
            <tr>
                <th scope="row">From</th>
                <td>{journey.from.full} ({journey.from.shortcode})</td>
            </tr>
            <tr>
                <th scope="row">To</th>
                <td>{journey.to.full} ({journey.to.shortcode})</td>
            </tr>
            <tr>
                <th scope="row">Via</th>
                <td>
                    {#if journey.via}
                        {#each journey.via as station, i}
                            {#if i !== 0}, {/if}{station.full} ({station.shortcode})
                        {/each}
                    {:else}
                        <span class="text-secondary"><i>n/a</i></span>
                    {/if}
                </td>
            </tr>
            <tr>
                <th scope="row">Date</th>
                <td>{formatDate(journey.date)}</td>
            </tr>
            <tr>
                <th scope="row">Distance</th>
                <td>{roundFloat(journey.distance, 2)} miles</td>
            </tr>
            {#if journey.returnID }
                <tr>
                    <th scope="row">Return</th>
                    <td><a href="#/journeys/{journey.returnID}" on:click={() => switchToJourney(journey.returnID)}>{journey.to.full} to {journey.from.full}</a></td>
                </tr>
            {/if}
            </tbody>
        </table>

        <div class="mb-4">
            <button class="btn btn-outline-danger" on:click={deleteSelf}>Delete this journey</button>
            {#if !journey.returnID }
                <button class="btn btn-outline-primary" on:click={createReturn}>Create return</button>
            {/if}
        </div>

        <p class="text-secondary">Journey ID: <code>{journey.id}</code></p>
    {/if}
</BaseLayout>