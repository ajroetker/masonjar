//TODO
$( function() {
    var tool = new paper.Tool();
    paper.setup('myCanvas');
    var center = new paper.Point(50, 50);
    var rectangleSize = new paper.Size(50, 75);
    var cornerSize = new paper.Size(rectangleSize.width / 7, rectangleSize.width / 7);
    var rectangle = new paper.Rectangle({
        center: center,
        size: rectangleSize,
        fillColor: 'white',
        strokeColor: 'black'
    });
    var specChars = {
        //spades
        1   : { spec: '\u2660', color: 'black' },
        //hearts
        2   : { spec: '\u2665', color: '#780000' },
        //clubs
        3    : { spec: '\u2663', color: 'black' },
        //diamonds
        4 : { spec: '\u2666', color: '#780000' },
    }

    function Board(object) {
        this.nertz  = object.nertz;
        this.stream = object.stream;
        this.lake   = object.lake;
        this.river  = object.river;
        this.hand   = object.hand;
        this.render = function() {
            var card;
            var center = new paper.Point(50, 50);
            for (var index in this.river) {
                card = new Card( river[index] );
                card.render(center);
                center.x += 75;
            };
            center.x += 25;
            for (var index in this.nertz) {
                card = new Card( nertz[index] );
                card.render(center);
                center.x -= .25;
                center.y -= .25;
            };

            center.y += 150;
            for (var index in this.stream) {
                card = new Card( stream[index] );
                card.render(center);
                center.x -= .25;
                center.y -= .25;
            };
            center.x -= 100;
            for (var index in this.hand) {
                card = new Card( hand[index] );
                card.render(center);
                if ( index < 3 ) {
                    center.x -= 10;
                    center.y -= 10;
                } else {
                    center.x -= .25;
                    center.y -= .25;
                };
            };
        }
    }

    function Card( suit, value, owner ) {
        this.suit  = suit;
        this.value = value;
        this.owner = owner;
        this.render = function(center) {
            var color, content;
            switch (value) {
                case 1:
                    content = 'A';
                    break;
                case 11:
                    content = 'J';
                    break;
                case 12:
                    content = 'Q';
                    break;
                case 13:
                    content = 'K';
                    break;
                default:
                    content = i;
                    break;
            }
            content += specChars[suit].spec;
            color = specChars[suit].color;

            rectangle.center = center;
            var path = new paper.Path.RoundRectangle(rectangle, cornerSize);
            path.fillColor = 'white';
            path.strokeColor = 'black';

            return new paper.Group({
                children: [
                    path,
                    new paper.PointText({
                        content: content,
                        fontSize: (16 * rectangleSize.width/50) + 'pt',
                        strokeColor: color,
                        fillColor: color,
                        justification: 'center',
                        position: path.position,
                    })]});
        };
    };

    var index;
    for (var i = 1; i <= 13; i++) {
        for (var j = 1; j <= 4; j++) {
            index = ( i - 1 ) * 4 + ( j - 1);
            var card = new Card(j, i, "me");
            if (index < 4) {
                card.render(center);
                if ( index === 3 ) {
                    center.x += 100;
                } else {
                    center.x += 75;
                }
            } else {
                if ( index < 13 ) {
                    card.render(center);
                    center.x -= .25;
                    center.y -= .25;
                    if ( index === 12 ) { center.y += 150 };
                } else {
                    card.render(center);
                    center.x -= .25;
                    center.y -= .25;
                }
            }
        };
    };

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

    tool.onMouseDown = onMouseDown;
    tool.onMouseUp = onMouseUp;
    tool.onMouseDrag = onMouseDrag;
    tool.activate();
});
