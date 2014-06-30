$( function() {
    paper.setup('myCanvas');

    var hitOptions = {
        segments: true,
        stroke: true,
        fill: true,
        tolerance: 0,
    };

    var x, y, hit, hitResult;
    function onMouseDown(event) {
        hitResult = paper.project.hitTest(event.point, hitOptions);
        if (!hitResult) {
            return;
        }
        hitResult = hitResult.item.parent;
        x = hitResult.position.x;
        y = hitResult.position.y;
        hitResult.bringToFront();
        hit = true;
    }

    function onMouseDrag(event) {
        if (!hit) {
            return;
        }
        hitResult.position.x += event.delta.x;
        hitResult.position.y += event.delta.y;
    }

    function onMouseUp(event) {
        if (!hit) {
            return;
        }
        hitResult.sendToBack();
        var other = paper.project.hitTest(event.point, hitOptions);
        other = other.item.parent;
        hitResult.bringToFront();
        if (!other || other === hitResult ) {
            hitResult.position.x = x;
            hitResult.position.y = y;
            return;
        }
        hitResult.position = other.position;
        hitResult.position.y += 10;
        hit = false;
    }

    var tool = new paper.Tool();
    tool.onMouseDown = onMouseDown;
    tool.onMouseUp = onMouseUp;
    tool.onMouseDrag = onMouseDrag;
    tool.activate();
});
