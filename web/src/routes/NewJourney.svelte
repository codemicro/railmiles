<script>
    import BaseLayout from "../components/BaseLayout.svelte"
    import {leftPad, makeURL} from "../util.js";
    import Loading from "../components/Loading.svelte";
    import {push} from "svelte-spa-router";
    import ErrorAlert from "../components/ErrorAlert.svelte";

    let problem
    let loading

    let inputs = {
        rawDate: undefined,
        date: undefined,
        route: undefined,
        manualDistance: undefined,
        isReturn: false,
    }

    $: inputs.manualDistance = parseFloat(inputs.manualDistance)
    $: {
        inputs.date = new Date(Date.parse(inputs.rawDate))
        console.log(inputs.rawDate)
    }

    const setDateToday = () => {
        const today = new Date(Date.now());
        inputs.rawDate = `${today.getFullYear()}-${leftPad(today.getMonth()+1, "0", 2)}-${leftPad(today.getDate(), "0", 2)}`
        console.log("set to", inputs.rawDate)
    }

    const doFormSubmit = async (event) => {
        event.preventDefault()

        if (!inputs.rawDate) {
            problem = "Please set the date of travel"
            return
        }

        if (!inputs.route) {
            problem = "Please provide a route"
            return
        }

        loading = true

        let response;
        try {
            response = await fetch(
                makeURL("/api/journeys"),
                {
                    method: "POST",
                    headers: {"Content-Type": "application/json"},
                    body: JSON.stringify(inputs),
                },
            )
        } catch (e) {
            alert(e.toString())
            loading = false
            return
        }

        const responseJSON = await response.json()

        loading = false

        if (response.status === 400) {
            problem = responseJSON.message
            return
        }

        await push(`/journeys/${responseJSON.id}`)
    }
</script>

<BaseLayout>
    {#if loading}
        <Loading transparent={true}/>
    {/if}

    <h1><i class="bi-plus-lg"></i> Log new journey</h1>
    <div class="pt-4"></div>

    {#if problem}
        <ErrorAlert message={problem} />
        <div class="pt-4"></div>
    {/if}

    <form on:submit={doFormSubmit}>
        <div class="row pb-2">
            <div class="col-sm">
                <label for="inputTravelDate" class="form-label">Date of travel <a class="link-primary"
                                                                                  on:click={setDateToday}>(today)</a></label>
            </div>
            <div class="col-sm-8">
                <input type="date" id="inputTravelDate" class="form-control" bind:value={inputs.rawDate}>
            </div>
        </div>

        <div class="row pb-2">
            <div class="col-sm">
                <label for="inputRoute" class="form-label">Route</label>
                <div class="form-text">Locations should be entered with the short code (eg: <code>SLY</code>) and
                    optionally the service UID (eg: <code>SLY, C16977</code>). Seperate locations with a newline.
                </div>
            </div>
            <div class="col-sm-8">
                <textarea type="date" id="inputRoute" rows="7" class="form-control"
                          bind:value={inputs.route}></textarea>
            </div>
        </div>

        <div class="row pb-2">
            <div class="col-sm">
                <label for="inputManualDistance" class="form-label">Manual distance</label>
                <div class="form-text">Leave blank to auto-detect. Enter values in miles.</div>
            </div>
            <div class="col-sm-8">
                <input type="number" step="any" id="inputManualDistance" class="form-control" placeholder="Auto-detect"
                       bind:value={inputs.manualDistance}>
            </div>
        </div>

        <div class="row pb-2">
            <div class="col-sm">
                <label for="inputReturnJourney" class="form-label">Was this a return journey?</label>
            </div>
            <div class="col-sm-8">
                <input type="checkbox" id="inputReturnJourney" class="form-check-input" bind:checked={inputs.isReturn}>
            </div>
        </div>

        <button type="submit" class="btn btn-primary">Submit</button>
    </form>
</BaseLayout>

<style>
    a.link-primary {
        cursor: pointer;
    }
</style>