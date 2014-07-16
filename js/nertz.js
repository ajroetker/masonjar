var cardChars = {
    1 : { spec: '\u2660', color: 'black', suit: 'spades' },
    2 : { spec: '\u2665', color: '#780000', suit: 'hearts' },
    3 : { spec: '\u2663', color: 'black', suit: 'clubs' },
    4 : { spec: '\u2666', color: '#780000', suit: 'diamonds' },
}

var values = {
    1: 'Ace',
    2: '2',
    3: '3',
    4: '4',
    5: '5',
    6: '6',
    7: '7',
    8: '8',
    9: '9',
    10: '10',
    11: 'Jack',
    12: 'Queen',
    13: 'King'
}

function cardImg(suit, value) {
    return $('<img style="z-index: 0;"' +
             'src="http://openclipart.org/people/' +
             'nicubunu/nicubunu_White_deck_' +
             value + '_of_' + suit + '.svg" />')
}

function Board(object) {
    this.nertz  = object.Nertz;
    this.stream = object.Stream;
    this.river  = object.River;
    this.show   = object.Show;
    this.render = function() {
        var pile = 0;
        $('#boardGame div img').remove();
        $.each( this.river, function( jndex, cards ) {
            $.each( cards, function( index, card ) {
                if ( card.Value ) {
                    img = cardImg( cardChars[card.Suit].suit, values[card.Value] )
                    $('#pile' + pile).append(img);
                    pile += 1;
                    img.droppable({
                        drop: function(event, ui) {
                            $( event.target ).css('z-index', 10);
                        }
                    });
                    img.draggable({
                        zIndex: 100,
                        stop: function(event, ui) {
                            $( event.target ).css('z-index', 100);
                        }
                    });
                }
            });
        });
        var position;
        $.each( this.nertz, function( index, card ) {
            if ( card.Value ) {
                img = cardImg( cardChars[card.Suit].suit, values[card.Value] )
                $('#nertz').append(img);
                position = img.position();
                if (index) {
                    img.css( 'top', position.top + .5 )
                       .css( 'left', position.left )
                       .css( 'position', 'absolute' );
                }
            }
        });
        $.each( this.stream, function( index, card ) {
            if ( card.Value ) {
                img = $('<img src="http://openclipart.org/people/nicubunu/nicubunu_Card_backs_simple_blue.svg" />');
                $('#stream').append(img);
                position = img.position();
                if (index) {
                    img.css( 'top', position.top )
                       .css( 'left', position.left - .5)
                       .css( 'position', 'absolute' );
                }
            }
        });
        $.each( this.show, function( index, card ) {
            if ( card.Value ) {
                img = cardImg( cardChars[card.Suit].suit, values[card.Value] )
                $('#show').append(img);
                img.draggable({
                    zIndex: 100,
                });
            }
        });

    }
}
