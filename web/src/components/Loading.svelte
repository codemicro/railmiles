<script>
    export let transparent = false
    export let text = undefined;
</script>

<div class="panel" class:transparent>
    <div class="d-flex gap-4 flex-column justify-content-center align-items-center">
        <div class="spinner-border" style="width: 4rem; height: 4rem;" role="status">
            <span class="visually-hidden">Loading...</span>
        </div>
        {#if text}
            <p>{text}</p>
        {/if}
    </div>
</div>

<style>
    div.panel {
        position: absolute;
        top: 0;
        left: 280px;
        width: calc(100% - 280px);
        height: 100%;
        background-color: #ffffff;
        z-index: 1500;
    }

    @media (max-width: 575.98px) {
        div.panel {
            left: 0;
            width: 100%;
        }
    }

    div.panel.transparent {
        background-color: rgba(255, 255, 255, 0.5);
    }

    div.panel > div {
        height: 100%;
    }
</style>