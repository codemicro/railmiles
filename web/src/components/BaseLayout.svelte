<script>
    import {location} from "svelte-spa-router"
    import sidebarOptions from "./sidebarOptions.js"
</script>

<div class="d-flex flex-column flex-shrink-0 p-3 text-white bg-dark sidebar d-none d-md-block">
    <a href="/" class="d-flex align-items-center mb-3 mb-md-0 me-md-auto text-white text-decoration-none">
        <i class="bi-train-front-fill pe-2" style="font-size: 32px;"></i>
        <span class="fs-4">RailMiles</span>
    </a>
    <hr>
    <ul class="nav nav-pills flex-column mb-auto">
        {#each sidebarOptions as {name, icon, path}}
            <li>
                <a href="#{path}" class={path === $location ? "nav-link active" : "nav-link text-white"}>
                    <i class="bi-{icon}" style="font-size: 16px;"></i>
                    {name}
                </a>
            </li>
        {/each}
    </ul>
</div>

<div class="navbar bg-dark text-white d-block d-md-none">
    <div class="container-fluid">
        <a href="/" class="navbar-brand d-flex align-items-center text-white text-decoration-none">
            <i class="bi-train-front-fill pe-2"></i>
            <span class="fs-4">RailMiles</span>
        </a>
        <button class="navbar-toggler text-white" type="button" data-bs-toggle="collapse" data-bs-target="#navMobileDropdown">
            <i style="font-size: 1.75rem" class="bi-list"></i>
        </button>
    </div>
</div>
<div class="collapse" id="navMobileDropdown" data-bs-theme="dark">
    <ul class="nav nav-pills flex-column mb-auto bg-dark p-2">
        {#each sidebarOptions as {name, icon, path}}
            <li>
                <a href="#{path}" class={path === $location ? "nav-link active" : "nav-link text-white"}>
                    <i class="bi-{icon}" style="font-size: 16px;"></i>
                    {name}
                </a>
            </li>
        {/each}
    </ul>
</div>

<div class="custom-container" style="">
    <div class="pt-5 d-none d-sm-block"></div>
    <div class="pt-4 d-block d-md-none"></div>
    <slot></slot>
    <div class="pt-5"></div>
</div>

<style>
    .sidebar {
        position: fixed;
        top: 0;
        bottom: 0;
        left: 0;
        box-shadow: 0 2px 5px 0 rgb(0 0 0 / 5%), 0 2px 10px 0 rgb(0 0 0 / 5%);
        width: 280px;
    }

    .custom-container {
        width: 80%;
        margin-left: auto;
        margin-right: auto;
        padding-left: 280px;
    }

    @media (max-width: 575.98px) {
        .custom-container {
            width: 90%;
            padding-left: unset;
        }
    }
</style>