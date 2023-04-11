

<div class="dashboard-container">
  <TimeSelector></TimeSelector>
  <div class="dashboard-content">
    <h2>Live Cluster View</h2>
    <p>View the current state of your cluster in real time.</p>
    <span>Selected Timeframe: {$selectedTimeframe.timeframe}</span>
    <div id="visualization"></div>
  </div>
</div>


<script lang="ts">
  import { onMount } from "svelte";
  import TimeSelector from "$lib/TimeSelector.svelte";
  import { updateDashboard } from "./Dashboard.svelte.ts";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";

  const selectedTimeframe = writable({timeframe: "1h", from: "", to: ""});

  setContext("selectedTimeframe", selectedTimeframe);

  // $:updateDashboard($selectedTimeframe.timeframe);

  onMount(() => {
      // const selectedTimeframe = $selectedTimeframe;
      updateDashboard($selectedTimeframe.timeframe);
  });
</script>


<style>
  .dashboard-container {
    padding-top: 4rem;
    font-family: "Arial", sans-serif;
    background-color: #f5f5f5;
    min-height: 100vh;
  }
  
  .dashboard-content {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
    background-color: #fff;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12), 0 1px 2px rgba(0, 0, 0, 0.24);
    border-radius: 4px;
  }
</style>