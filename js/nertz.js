var cardChars = {
    1 : { spec: '\u2660', color: 'black' }, //spades
    2 : { spec: '\u2665', color: '#780000' }, //hearts
    3 : { spec: '\u2663', color: 'black' }, //clubs
    4 : { spec: '\u2666', color: '#780000' }, //diamonds
}

var rectangleSize = new paper.Size(50, 75);
var cornerSize    = new paper.Size(rectangleSize.width / 7, rectangleSize.width / 7);

var rectangle = new paper.Rectangle({
    size: rectangleSize,
    fillColor: 'white',
    strokeColor: 'black'
});

function Board(object) {
    this.nertz  = object.Nertz;
    this.stream = object.Stream;
    this.river  = object.River;
    this.show   = object.Show;

    this.render = function() {
        var card;
        var center = new paper.Point(50, 50);
        for (var index in this.river) {
            for (var jndex in this.river[index]) {
                card = new Card( this.river[index][jndex] );
                card.render(center);
                center.x += 75;
            };
        };
        center.x += 25;
        for (var index in this.nertz) {
            card = new Card( this.nertz[index] );
            card.render(center);
            center.x -= .25;
            center.y -= .25;
        };

        center.y += 150;
        for (var index in this.stream) {
            card = new Card( this.stream[index] );
            card.render(center);
            center.x -= .25;
            center.y -= .25;
        };
        center.x -= 100;
        if (this.length > 0) {
            for (var index in this.show) {
                card = new Card( this.show[index] );
                card.render(center);
                center.x -= .25;
                center.y -= .25;
            };
        }
    }
}

function Card(object) {
    this.suit  = object.Suit;
    this.value = object.Value;
    this.owner = object.Owner;
    this.render = function(center) {
        var content = cardChars[this.suit].spec;
        switch (this.value) {
            case 1:
                content = 'A' + content;
                break;
            case 11:
                content = 'J' + content;
                break;
            case 12:
                content = 'Q' + content;
                break;
            case 13:
                content = 'K' + content;
                break;
            default:
                content = this.value + content;
                break;
        }

        var color = cardChars[this.suit].color;

        rectangle.center = center;
        var path = new paper.Path.RoundRectangle(rectangle, cornerSize);
        path.fillColor = 'white';
        path.strokeColor = 'black';

        var text = new paper.PointText({
            content       : content,
            fontSize      : (16 * rectangleSize.width/50) + 'pt',
            strokeColor   : color,
            fillColor     : color,
            justification : 'center',
            position      : path.position,
        })

        return new paper.Group({
            children: [ path, text ]
        });
    };
};
