<!-- Heading.svelte -->
<style>
    :global(body) {
      margin: 0;
      padding: 0;
    }
  
    .heading {
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      font-size: 2rem;
      font-weight: bold;
      color: #34495e;
      text-align: center;
      margin: 0;
      padding: 1rem 0;
      width: 100%;
      background-color: #1abc9c;
      box-shadow: 0px 2px 5px rgba(0, 0, 0, 0.1);
      position: fixed;
      top: 0;
      left: 0;
      z-index: 1000;
    }
  </style>
  
  <h1 class="heading"><slot></slot></h1>
  