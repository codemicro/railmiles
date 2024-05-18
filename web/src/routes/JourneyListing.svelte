<script>
    import BaseLayout from "../components/BaseLayout.svelte"
    import JourneyTable from "../components/JourneyTable.svelte"
    import {onMount} from "svelte"
    import Loading from "../components/Loading.svelte"
    import {makeURL} from "../util.js"

    let journeys = []
    let totalNumPages
    let currentPage = 0
    let ready = false
    let transparentLoading = false

    const getPage = async (n) => {
        let response;
        try {
            response = await fetch(makeURL("/api/journeys?page=" + n));
        } catch (e) {
            alert(e.toString())
            return
        }
        return await response.json()
    }

    const makeWindow = () => {
        let parts

        if (totalNumPages < 3) {
            parts = []
            for (let i = 0; i < totalNumPages; i += 1) {
                parts.push(i)
            }
            return parts
        }

        const cursor = currentPage + 1

        if (cursor - 1 == 0) {
            // we've got the first one selected
            parts = [cursor, cursor + 1, cursor + 2]
        } else if (cursor + 1 > totalNumPages) {
            // we've got the last one selected
            parts = [cursor - 2, cursor - 1, cursor]
        } else {
            parts = [cursor - 1, cursor, cursor + 1]
        }

        return parts
    }

    onMount(async () => {
        const resp = await getPage(currentPage)
        totalNumPages = resp.numPages
        journeys = resp.data
        ready = true
        transparentLoading = true
    })

    $: {
        ready = false
        getPage(currentPage).then((x) => {
            journeys = x.data
            ready = true
        })
    }
</script>

<BaseLayout>
    {#if !ready}
        <Loading transparent={transparentLoading}/>
    {/if}

    <h1><i class="bi-table"></i> Journey listing</h1>

    <div class="pt-4"></div>

    <nav class="d-flex justify-content-center">
        <ul class="pagination">
            <li class={currentPage === 0 ? "page-item disabled" : "page-item"}><a role="button" tabindex="0" class="page-link" on:click={() => {currentPage = 0}}><i class="bi-chevron-double-left"></i></a></li>
            <li class={currentPage === 0 ? "page-item disabled" : "page-item"}><a role="button" tabindex="0" class="page-link" on:click={() => {currentPage--}}><i class="bi-chevron-left"></i></a></li>
            {#each makeWindow() as pageNumber, _ (pageNumber)}
                <li class={currentPage === pageNumber - 1 ? "page-item active" : "page-item"}><a role="button" tabindex="0" class="page-link" on:click={() => {currentPage=pageNumber - 1}}>{pageNumber}</a></li>
            {/each}
            <li class={currentPage+1 === totalNumPages ? "page-item disabled" : "page-item"}><a role="button" tabindex="0" class="page-link" on:click={() => {currentPage++}}><i class="bi-chevron-right"></i></a></li>
            <li class={currentPage+1 === totalNumPages ? "page-item disabled" : "page-item"}><a role="button" tabindex="0" class="page-link" on:click={() => {currentPage = totalNumPages-1}}><i class="bi-chevron-double-right"></i></a></li>
        </ul>
    </nav>

    <JourneyTable journeys={journeys}/>
</BaseLayout>