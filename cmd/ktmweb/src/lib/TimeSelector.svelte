<!-- TimeframeSelector.svelte -->
<div class="timeframe-selector">
    <label for="timeframe">Timeframe</label>
    <select id="timeframe" bind:value={timeframe} on:change={handleTimeframeChange}>
      <option value="1h">1 hour</option>
      <option value="2h">2 hours</option>
      <option value="4h">4 hours</option>
      <option value="8h">8 hours</option>
      <option value="12h">12 hours</option>
      <option value="24h">24 hours</option>
      <option value="48h">48 hours</option>
      <option value="72h">72 hours</option>
      <option value="168h">1 week</option>
      <option value="336h">2 weeks</option>
      <option value="custom">Custom</option>
    </select>
    {#if timeframe === 'custom'}
    <div>
      <label for="from">From:</label>
      <input type="datetime-local" id="from" bind:value={from}>
      <label for="to">To:</label>
      <input type="datetime-local" id="to" bind:value={to}>
    </div>
    {/if}
  </div>

<script>
    import { onMount } from "svelte";
    import { getContext } from "svelte";
    import { writable } from "svelte/store";
	import { setContext } from "svelte";

    export let timeframe = "1h";
    export let from = "";
    export let to = "";

    const selectedTimeframe = getContext('selectedTimeframe');

    // $: selectedTimeframe.set({timeframe, from, to});


    function handleTimeframeChange() {
        $selectedTimeframe = {timeframe, from, to};
        if (timeframe !== 'custom') {
            console.log("Relative timeframe selected: " + timeframe);
        } else {
            console.log("Absolute timeframe selected: From " + from + " to " + to);
        }
    }

    onMount(() => {
        // $timeframe && console.log("Timeframe: " + $timeframe);
    });
</script>


<style>
    .timeframe-selector {
      display: flex;
      align-items: center;
      justify-content: flex-end;
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      background-color: #2c3e50;
      color: white;
      padding: 0.5rem;
      padding-top: 1.5rem;
    }
  
    .timeframe-selector label {
      margin-right: 0.5rem;
      font-size: 0.9rem;
    }
  
    .timeframe-selector select,
    .timeframe-selector input {
      font-family: inherit;
      font-size: 0.9rem;
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      border: 1px solid #34495e;
      color: #34495e;
      background-color: white;
      margin-right: 0.5rem;
    }
  
    .timeframe-selector select:focus,
    .timeframe-selector input:focus {
      outline: none;
      border-color: #1abc9c;
    }
  </style>