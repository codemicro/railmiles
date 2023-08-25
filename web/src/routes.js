import Home from './routes/Home.svelte'
import JourneyListing from './routes/JourneyListing.svelte'
import NewJourney from "./routes/NewJourney.svelte";
import NotFound from './routes/NotFound.svelte'
import JourneyDetail from "./routes/JourneyDetail.svelte";

export default {
    '/': Home,
    '/journeys': JourneyListing,
    '/journeys/:id': JourneyDetail,
    '/new': NewJourney,
    '*': NotFound,
}