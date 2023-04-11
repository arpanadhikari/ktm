
import * as d3 from "d3";
import type { SvelteComponent } from "svelte";

export interface DashboardProps {
    title: string;
  }
  
//   export class Dashboard extends SvelteComponent {
//     constructor(options: { target: Element; props: DashboardProps }) {
//       super(options);
//     }
//   }

// document.addEventListener("DOMContentLoaded", () => {
//     updateDashboard();
//     // setInterval(updateDashboard, 2000);
//   });
  
export  async function mapPodsToNodes() {
    const clusterSnapshot = await getClusterSnapshot("2h");
    const nodeData = clusterSnapshot.data[1];
    const podData = clusterSnapshot.data[0];
  
    console.log(nodeData);
    console.log(podData);
  
    const groupedPods = d3.group(podData.children, d => d.nodename);
    nodeData.children.forEach(node => {
      node.children = groupedPods.get(node.name) || [];
    });
  
    return nodeData;
  }
  
export  function customTile(node, x0, y0, x1, y1) {
    if (node.children) {
      return d3.treemapSquarify(node, x0, y0, x1, y1);
    }
  
    const ratio = node.data.size.memory * 10 / (node.data.size.cpu * 20);
    const width = x1 - x0;
    const height = y1 - y0;
    const dx = width > height / ratio ? (width - height / ratio) / 2 : 0;
    const dy = width <= height / ratio ? (height - width * ratio) / 2 : 0;
  
    return [x0 + dx, y0 + dy, x1 - dx, y1 - dy];
  }
  
export  function computeSize(node) {
    return node.children
      ? node.children.reduce((acc, child) => acc + computeSize(child), 0)
      : node.size.cpu + node.size.memory + 10;
  }
  
export  async function getClusterSnapshot(relativeTime) {
    const response = await fetch(`http://localhost:8080/clustersnapshot?relativeTime=${relativeTime}`);
    return await response.json();
  }
  
  export  async function updateDashboard(timeframe) {
    const nodeData = await mapPodsToNodes();
  
    const svg = d3
      .select("#visualization")
      .selectAll("svg")
      .data([nodeData])
      .join("svg")
      .attr("width", window.innerWidth)
      .attr("height", window.innerHeight);
  
    const root = d3
      .hierarchy(nodeData)
      .sum(d => computeSize(d))
      .sort();
  
    const treemap = d3
      .treemap()
      .size([800, 800])
      .paddingInner(10)
      .paddingTop(5)
      .paddingRight(2)
      .paddingBottom(2)
      .paddingLeft(2)
      .round(2)
      .tile(d3.treemapSquarify.ratio(1))(root);
  
    const g = svg
      .selectAll("g")
      .data([treemap])
      .join("g")
      .attr("transform", `translate(${treemap.leaves()[0].x0},${treemap.leaves()[0].y0})`);
  
    const nodes = g
      .selectAll("a")
      .data(treemap.descendants())
      // .join("a")
      .join(
        enter => enter.append("a").classed("node-enter", true),
        update => update.classed("node-update", true),
        exit => exit.classed("node-exit", true).remove()
      )
      .attr("href", d => {
        return d.height === 1
          ? `nodehistory/${d.data.name}?relativeTime=${timeframe}`
          : `podhistory/${d.data.name}?relativeTime=${timeframe}`;
      });
  
    nodes
      .selectAll("rect")
      .data(d => [d])
      // .join("rect")
      .join(
        enter => enter.append("rect").classed("node-enter", true),
        update => update.classed("node-update", true),
        exit => exit.classed("node-exit", true).remove()
      )
      .attr("x", d => d.x0)
      .attr("y", d => d.y0)
      .attr("width", d => d.x1 - d.x0)
      .attr("height", d => d.y1 - d.y0)
      .style("fill", d => (d.height === 0 ? "#788b9d": (d.height === 1 ? "#03385c" : "#3568a3")));
  
    nodes
      .selectAll("title")
      .data(d => [d])
      .join("title")
      .text(d => d.data.name);
  
    nodes
      .selectAll("text")
      .data(d => [d])
      .join("text")
      .attr("x", d => d.x0 + 10)
      .attr("y", d => (d.height === 1 ? d.y0 + 20 : d.y0 + 40))
      .attr("font-size", "14px")
      .attr("fill", "white");
      // .text(d => d.data.name);
  }
