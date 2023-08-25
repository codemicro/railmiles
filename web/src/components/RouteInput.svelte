<script>
    export let route = [["", ""], ["", ""]]

    const removeByIndex = (event, idx) => {
        event.preventDefault()
        console.log("before", JSON.stringify(route))
        route.splice(idx, 1)
        route = [...route.slice(0, idx), ...route.slice(idx)]
        console.log("trigger", JSON.stringify(route))
    }

    const addAtIndex = (event, idx) => {
        event.preventDefault()
        route = [...route.slice(0, idx), ["", ""], ...route.slice(idx, route.length)]
    }
</script>

{#each route as row, i}
    <div class="input-group pb-1">
        <input type="text" class="form-control" placeholder="Station" bind:value={route[i][0]}>
        <input type="text" class="form-control" placeholder="Service UID"
               bind:value={route[i][1]}>
        <button class="btn btn-sm btn-primary" on:click={(e) => addAtIndex(e, i+1)}>
            <i class="bi-plus-lg"></i>
        </button>
        <button class="btn btn-sm btn-danger" disabled={i === route.length - 1 || i === 0}
                on:click={(e) => removeByIndex(e, i)}>
            <i class="bi-trash3-fill"></i>
        </button>
    </div>
{/each}
