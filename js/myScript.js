//TODO
var center = new Point(50, 50);
var rectangleSize = new Size(50, 75);
var cornerSize = new Size(rectangleSize.width / 10, rectangleSize.width / 10);
var rectangle = new Rectangle({
    center: center,
    size: rectangleSize,
    fillColor: 'white',
    strokeColor: 'black'
});

var path, cardText, rectangle;
function renderCard(content, center, color) {
    rectangle.center = center
    path = new Path.RoundRectangle(rectangle, cornerSize);
    path.fillColor = 'white';
    path.strokeColor = 'black';

    return new Group({
        children: [
            path,
            new PointText({
                content: content,
                fontSize: (16 * rectangleSize.width/50) + 'pt',
                strokeColor: color,
                fillColor: color,
                justification: 'center',
                position: path.position,
            }),
            new PointText({
                content: content,
                fontSize: (16 * rectangleSize.width / 100) + 'pt',
                strokeColor: color,
                fillColor: color,
                justification: 'center',
                position: path.position -  new Point( rectangleSize.width / 4, 11 * rectangleSize.height / 30 ),
            }),
            new PointText({
                content: content,
                fontSize: (16 * rectangleSize.width / 100) + 'pt',
                strokeColor: color,
                fillColor: color,
                justification: 'center',
                position: path.position + new Point( rectangleSize.width / 4, 11 * rectangleSize.height / 30 ),
                rotation: 180,
            })
        ],
    });
}

var specChars = {
    spades   : { spec: '\u2660', color: 'black' },
    hearts   : { spec: '\u2665', color: 'red' },
    clubs    : { spec: '\u2663', color: 'black' },
    diamonds : { spec: '\u2666', color: 'red' },
}

var cardText, color;
for (var i = 1; i <= 13; i++) {
    for (var specChar in specChars) {
        switch (i) {
            case 1:
                cardText = 'A';
                break;
            case 11:
                cardText = 'J';
                break;
            case 12:
                cardText = 'Q';
                break;
            case 13:
                cardText = 'K';
                break;
            default:
                cardText = i;
                break;
        }
        cardText += specChars[specChar].spec;
        color = specChars[specChar].color;
        renderCard(cardText, center, color);
        center.y += 50;
    }
    center.y = 50;
    center.x += 25;
}

var hitOptions = {
    segments: true,
    stroke: true,
    fill: true,
    tolerance: 5,
};

var hit, hitResult;
function onMouseDown(event) {
    hitResult = project.hitTest(event.point, hitOptions);
    if (!hitResult) {
        return;
    }
    hitResult = hitResult.item.parent;
    hitResult.bringToFront();
    hit = true;
}

function onMouseDrag(event) {
    if (!hit) {
        return;
    }
    hitResult.position +=  event.delta;
}

function onMouseUp(event) {
    if (!hit) {
        return;
    }
    hit = false;
}

var board = {
};
