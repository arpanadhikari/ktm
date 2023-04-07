document.addEventListener("DOMContentLoaded", () => {

    // create nodes on svg based on getNodes() function
    mapPodsToNodes().then(nodeData=>{
        // create svg
        const svg = d3.select("#visualization")
        .append("svg")
        .attr("width", window.innerWidth)
        .attr("height", window.innerHeight);
        
        console.log(nodeData)
        // create a simple root node
        const root = d3.hierarchy(nodeData)
        .sum(d => computeSize(d))
        .sort();
        
        // print root node
    console.log("Root node",root)

    // create a treemap with nodeData
    const treemap = d3.treemap()
    // .title("kubernetes time machine")
    // window.innerWidth, window.innerHeight
    .size([600,600])
        .paddingInner(10)
        .paddingTop(2)
        .paddingRight(2)
        .paddingBottom(2)
        .paddingLeft(2)
        .round(2)
        // .padding(0)
        .tile(d3.treemapSquarify.ratio(1))(root)

        console.log("Treemap",treemap)
        // console.log(treemap.leaves())    
        
    const g = svg.append("g")
    .attr("transform", "translate(" + treemap.leaves()[0].x0 + "," + treemap.leaves()[0].y0 + ")");
    

    

    // add nodes
    g.selectAll("a")
    .data(treemap.descendants().slice(0))
    .enter()
    .append("a")
    .attr("href", (d) => {
      if (d.depth === 0) {
        return "#"
      }
      return d.height === 1 ? "node/"+d.data.name : "pod/"+d.data.name
      return d.data.name+"/"
    
    })

    // add rectangles for each pod
    g.selectAll("a")
    .data(treemap.descendants().slice(0))
    .append("rect")
    .attr("x", (d) => {
        // console.log(d.x0, d.x1);
        // console.log(d)
        return d.x0-0;
    })
    .attr("y", (d) => {
        // console.log(d.y0, d.y1);
        return d.y0-0;
    })
    .attr("width", (d) => (d.x1 - d.x0)+0)
    .attr("height", (d) => (d.y1 - d.y0)+0)
    .style("fill", (d) => {
        // console.log("DATA type:", d);
        if (d.depth === 0) {
            return "lightblue"
        }
        return d.height === 1 ? "#03385c" : "#3568a3"
    })
    .attr("margin", (d) => {
        if (d.depth === 0) {
            return "0px"
        }
        return d.height === 1 ? "5px" : "10px"
    });
    
    // add text for each pod
    // g.selectAll("a")
    // .data(treemap.descendants().slice(0))
    // // .enter()
    // .append("text")
    // .attr("x", (d) => d.x0 + 10)
    // .attr("y", (d) => {
    //     if (d.depth === 0) {
    //         return "0"
    //     }
    //     return d.height === 1 ? d.y0 + 20 : d.y0 + 40
    // })
    // .attr("font-size", "14px")
    // .attr("fill", "white")
    // .text((d) => d.data.name)
    
    // print final treemap
    console.log("Final treemap",treemap)

    // add a title for each pod
    g.selectAll("a")
    .data(treemap.descendants().slice(0))
    // .enter()
    .append("title")
    .text((d) => d.data.name);

    g.selectAll("a")
    .data(treemap.descendants().slice(0))
    // .enter()
    .append("label")
    .text((d) => d.data.name);


    // g.append("text")
    //     .attr("clip-path", (d, i) => `url(${new URL(`#${uid}-clip-${i}`, location)})`)
    //   .selectAll("tspan")
    //   .data(treemap.descendants().slice(0))
    //   .join("tspan")
    //     .attr("x", 3)
    //     .attr("y", (d, i, D) => `${(i === D.length - 1) * 0.3 + 1.1 + i * 0.9}em`)
    //     .attr("fill-opacity", (d, i, D) => i === D.length - 1 ? 0.7 : null)
    //     .text(d => d.data.name);

    });



});
// function to map pods to nodes
async function mapPodsToNodes() {
    const nodeData = await getNodes();
    const podData = await getPods();
  
    // Use d3.group to group pods by nodeName
    const groupedPods = d3.group(podData.children, d => d.nodename);

    // Loop through nodes and add pods to each node's children array
    nodeData.children.forEach(node => {
      const nodePods = groupedPods.get(node.name);
      if (nodePods) {
        node.children = nodePods;
      } else {
        node.children = [];
      }
    });
  
    return nodeData;
  }
  
// function to create custom tiling
function customTile(node, x0, y0, x1, y1) {

  if (node.children) {
    return d3.treemapSquarify(node, x0, y0, x1, y1);
  }

  const ratio = node.data.size.memory*10 / node.data.size.cpu*20; // prioritize memory over CPU
  const width = x1 - x0;
  const height = y1 - y0;

  if (width > height / ratio) {
    const dx = (width - height / ratio) / 2;
    return [x0 + dx, y0, x1 - dx, y1];
  } else {
    const dy = (height - width * ratio) / 2;
    return [x0, y0 + dy, x1, y1 - dy];
  }
}

function computeSize(node) {
  if (!node.children) {
    // check if node has a pod
    console.log("Computesize cpu: "+node.size.cpu, "memory: "+node.size.memory, "size: "+(node.size.cpu+node.size.memory)+10)
    return (node.size.cpu+node.size.memory+10); // use CPU size as leaf node size
  } else {
    console.log("Computesize else")
    return node.children.reduce((acc, child) => acc + computeSize(child), 0);
    // compute size of internal node as sum of child sizes
  }
}

// function to fetch number of nodes
async function getNodes() {
    const response = await fetch('/nodehistory');
    const data = await response.json();
    console.log("Response:", data);
    return data;
}

async function getPods() {
  const response = await fetch('/podhistory');
  const data = await response.json();
  console.log("Response:", data);
  return data;
}