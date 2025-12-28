// Compute dynamic size
const containerId = 'viewPanel';
const container = document.getElementById(containerId);
const rect = container ? container.getBoundingClientRect() : { width: window.innerWidth, height: window.innerHeight };

const width = Math.max(320, Math.floor(rect.width || window.innerWidth));
const height = Math.max(240, Math.floor(rect.height || Math.floor(window.innerHeight * 0.6)));

const gridSize= 35;
const gridDotSize= 2;
const gridColor= '#e7e7e7';
const cx= width * 0.5;
const cy= height * 0.5;
const radius= Math.min(width, height) / 2 - 20;

function updateGrid(svg, zoomEvent) {
    svg.select('#dot-pattern')
        .attr('x', zoomEvent.transform.x)
        .attr('y', zoomEvent.transform.y)
        .attr('width', gridSize * zoomEvent.transform.k)
        .attr('height', gridSize * zoomEvent.transform.k)
        .selectAll('rect')
        .attr('x', (gridSize * zoomEvent.transform.k / 2) - (gridDotSize / 2))
        .attr('y', (gridSize * zoomEvent.transform.k / 2) - (gridDotSize / 2))
        .attr('opacity', Math.min(zoomEvent.transform.k, 1)); // Lower opacity as the pattern gets more dense.
}

export function drawTree(data, callback) {
    // Create a radial tree layout. The layout’s first dimension (x)
    // is the angle, while the second (y) is the radius.

    const tree = d3
        .tree()
        .size([2 * Math.PI, radius])
        .separation((a, b) => (a.parent === b.parent ? 0.25 : 2) / a.depth);

    // Sort the tree and apply the layout.
    const root = d3
        .hierarchy(data)
        .sort((a, b) => d3.ascending(a.data.name, b.data.name));

    const diagonal = d3
        .linkRadial()
        .angle((d) => d.x)
        .radius((d) => d.y);

    // Creates the SVG container.
    const svg = d3
        .create("svg")
        .attr("width", width)
        .attr("height", height)
        .attr("viewBox", [-cx, -cy, width, height])
        .attr("class", "tree");

    svg.append('pattern')
        .attr('id', 'dot-pattern')
        .attr('patternUnits', 'userSpaceOnUse')
        .attr('x', -cx)
        .attr('y', -cy)
        .attr('width', gridSize)
        .attr('height', gridSize)
        .append('rect')
        .attr('width', gridDotSize)
        .attr('height', gridDotSize)
        .attr('fill', gridColor)
        .attr('x', (gridSize / 2) - (gridDotSize / 2))
        .attr('y', (gridSize / 2) - (gridDotSize / 2));

    svg.append('rect')
        .attr('x', -cx)
        .attr('y', -cy)
        .attr('fill', 'url(#dot-pattern)')
        .attr('width', '100%')
        .attr('height', '100%');

    var content = svg.append('g').attr('id', 'content');

    // Append links.
    const gLink = content
        .append("g")
        .attr("fill", "none")
        .attr("stroke", "#2e6fde")
        .attr("stroke-opacity", 0.75)
        .attr("stroke-width",2.5);

    // Append nodes.
    const gNode = content
        .append("g")
        .attr("cursor", "pointer")
        .attr("pointer-events", "all");

    function update(event, source) {
        const nodes = root.descendants().reverse();
        const links = root.links();

        // Dynamically adjust the radius based on the max depth of visible nodes
        const maxVisibleDepth = d3.max(nodes, d => d.depth) || 1;
        tree.size([2 * Math.PI, maxVisibleDepth * 150]); // Adjust 100 to control spacing

        tree(root);

        // Compute the new tree layout.
        let left = root;
        let right = root;

        root.eachBefore((node) => {
            if (node.x < left.x) left = node;
            if (node.x > right.x) right = node;
        });

        const transition = svg
            .transition()
            .duration(190)
            .attr("height", height);

        // Update the nodes…
        const node = gNode.selectAll("g").data(nodes, (d) => d.id || (d.id = ++i));

        // Enter any new nodes at the parent's previous position.
        const nodeEnter = node
            .enter()
            .append("g")
            .attr('id', d => d.name)
            .attr(
                "transform",
                (d) =>
                    `rotate(${(source.x0 * 180) / Math.PI - 90}) translate(${
                        source.y0
                    },0)`
            )
            .attr("fill-opacity", 0)
            .attr("stroke-opacity", 0)
            .on("click", (event, d) => {
                if (d.depth != 0) {
                    d.children = d.children ? null : d._children;
                    update(event, d);
                }
                if ((!d._children || d.children) || (d.data["ref"] !== undefined)) {
                    callback(d.data)
                }
            })
            .on('mouseover', function(event, d) {
                d3.select(this).select('text')
                    .transition().duration(50)
                    .attr('font-weight', '800');
            })
            .on('mouseout', function(event, d) {
                d3.select(this).select('text')
                    .transition().duration(50)
                    .attr('font-weight', null);
            });

        nodeEnter
            .filter((d) => d.depth > 0)
            .filter((d) => d.data["ref"] !== undefined)
            .append("circle")
            .attr("r", 3.5)
            .attr("fill", "rgba(39,217,11,0.85)")
            .attr("stroke-width", 10);

        nodeEnter
            .filter((d)=> d.data["ref"] === undefined)
            .filter((d) => d.depth > 0)
            .append("rect")
            .attr("width", 10)
            .attr("height", 10)
            .attr("x", -5)
            .attr("y", -5)
            .attr("fill", "rgb(0,121,202)")

        nodeEnter
            .filter((d) => d.depth > 0)
            .append("text")
            .attr("font-size",20)
            .attr("dy", "0.1em")
            //start and end exchange after half a circle
            .attr("text-anchor", (d) => (d.x < Math.PI ? "start" : "end"))
            .attr(
                "transform",
                (d) =>
                    `rotate(${d.x >= Math.PI ? 180 : 0}) translate(${
                        d.x < Math.PI ? 8 : -8
                    })`
            )
            .text((d) => d.data.name)
            .attr("fill", "var(--text-color)")
            .attr("stroke-linejoin", "round")
            .attr("stroke-width", 2.5)
            .attr("stroke", "black")
            .attr("paint-order", "stroke");

        const nodeUpdate = node.merge(nodeEnter);

        // Update text attributes for both entering and updating nodes
        nodeUpdate.select("text")
            .attr("text-anchor", (d) => (d.x < Math.PI ? "start" : "end"))
            .attr(
                "transform",
                (d) =>
                    `rotate(${d.x >= Math.PI ? 180 : 0}) translate(${
                        d.x < Math.PI ? 8 : -8
                    })`
            );

        // Transition nodes to their new position.
        nodeUpdate
            .transition(transition)
            .attr(
                "transform",
                (d) => `rotate(${(d.x * 180) / Math.PI - 90}) translate(${d.y},0)`
            )
            .attr("fill-opacity", 1)
            .attr("stroke-opacity", 1);

        // Transition exiting nodes to the parent's new position.
        const nodeExit = node
            .exit()
            .transition(transition)
            .remove()
            .attr(
                "transform",
                (d) =>
                    `rotate(${(source.x * 180) / Math.PI - 90}) translate(${source.y},0)`
            )
            .attr("fill-opacity", 0)
            .attr("stroke-opacity", 0);

        // Update the links…
        const link = gLink.selectAll("path").data(links, (d) => d.target.id);

        // Enter any new links at the parent's previous position.
        const linkEnter = link
            .enter()
            .append("path")
            .attr("d", (d) => {
                const o = {x: source.x0, y: source.y0};
                return diagonal({source: o, target: o});
            });

        // Transition links to their new position.
        link.merge(linkEnter).transition(transition).attr("d", diagonal);

        // Transition exiting nodes to the parent's new position.
        link
            .exit()
            .transition(transition)
            .remove()
            .attr("d", (d) => {
                const o = {x: source.x, y: source.y};
                return diagonal({source: o, target: o});
            });

        // Stash the old positions for transition.
        root.eachBefore((d) => {
            d.x0 = d.x;
            d.y0 = d.y;
        });
    }

    let i = 0;
    root.x0 = 0;
    root.y0 = 0;

    root.eachBefore((d) => {
        d.x0 = d.x;
        d.y0 = d.y;
        // Stash children initially for collapsibility
        if (d.children) d._children = d.children;
    });

    if (root.children) {
        root.children = root._children;
        root._children = null;
    }

    root.each(d => {
        if (d.depth > 1 && d.children) {
            d._children = d.children;
            d.children = null;
        }
    });

    svg.call(d3.zoom()
        .scaleExtent([0.25, 2])
        .on("zoom", (event)=> {
            content.attr('transform', event.transform);
            updateGrid(svg, event); // We need to update the grid with every zoom event.
        }));

    update(null, root);
    return svg.node();
}