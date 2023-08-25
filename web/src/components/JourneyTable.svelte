<script>
    import {formatDate, roundFloat} from "../util.js";

    export let journeys = [];
    export let showMore = false;
</script>

<table class="table table-sm table-hover">
    <thead>
    <tr>
        <th scope="col">Date</th>
        <th scope="col">Route</th>
        <!--        <th scope="col">To</th>-->
        <th scope="col"></th>
        <th scope="col">Distance</th>
        <th scope="col">Return</th>
        <th scope="col"></th>
    </tr>
    </thead>
    <tbody>
    {#each journeys as journey (journey.id)}
        <tr>
            <td>{formatDate(journey.date)}</td>
            <td>{journey.from.full} to {journey.to.full}</td>
            <td>
                {#if journey.via}
                    via
                    {#each journey.via as station, i}
                        {#if i !== 0}, {/if}<abbr title="{station.full}">{station.shortcode}</abbr>
                    {/each}
                {/if}
            </td>
            <td>{roundFloat(journey.distance, 1)} miles</td>
            <td>
                {#if journey.return }<i class="bi-check-lg"></i>{/if}
            </td>
            <td><a href="#/journeys/{journey.id}"><i class="bi-three-dots"></i></a></td>
        </tr>
    {:else}
        <tr>
            <td colspan="6" class="text-center bg-warning-subtle text-warning-emphasis">Nothing to display!</td>
        </tr>
    {/each}
    {#if showMore}
        <tr>
            <td colspan="6"><a href="#/journeys">See more...</a></td>
        </tr>
    {/if}
    </tbody>

</table>

<style>
    table {
        overflow: scroll;
    }
</style>