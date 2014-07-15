var cardChars = {
    1 : { spec: '\u2660', color: 'black', suit: 'spades' },
    2 : { spec: '\u2665', color: '#780000', suit: 'hearts' },
    3 : { spec: '\u2663', color: 'black', suit: 'clubs' },
    4 : { spec: '\u2666', color: '#780000', suit: 'diamonds' },
}
var values = { 1: 'Ace', 11: 'Jack', 12: 'Queen', 13: 'King' }

function Board(object) {
    this.nertz  = object.Nertz;
    this.stream = object.Stream;
    this.river  = object.River;
    this.show   = object.Show;
    this.render = function() {
        var position;
        var pile = 0;
        $('#boardGame div div img').remove();
        $.each( this.river, function( jndex, cards ) {
            $.each( cards, function( index, card ) {
                if ( card.Value ) {
                    var value = values[card.Value];
                    if ( !value ) {
                        value = card.Value;
                    }
                    var src = "/nicubunu_White_deck_" + value + "_of_" + cardChars[card.Suit].suit + ".svg";
                    img = $('<img style="z-index: 0;" src="http://openclipart.org/people/nicubunu' + src + '" />');
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
        $.each( this.nertz, function( index, card ) {
            if ( card.Value ) {
                var value = values[card.Value];
                if ( !value ) {
                    value = card.Value;
                }
                var src = "/nicubunu_White_deck_" + value + "_of_" + cardChars[card.Suit].suit + ".svg";
                img = $('<img src="http://openclipart.org/people/nicubunu' + src + '" />');
                $('#nertz').append(img);
                if (index) {
                    var offset;
                    img.css( 'top', position.top + .5 ).css( 'left', position.left ).css( 'position', 'absolute' );
                }
                position = img.position();
            }
        });
        $.each( this.stream, function( index, card ) {
            if ( card.Value ) {
                img = $('<img src="http://openclipart.org/people/nicubunu/nicubunu_Card_backs_simple_blue.svg" />');
                $('#stream').append(img);
                if (index) {
                    var offset;
                    img.css( 'top', position.top ).css( 'left', position.left - .5).css( 'position', 'absolute' );
                }
                position = img.position();
            }
        });
        $.each( this.show, function( index, card ) {
            if ( card.Value ) {
                var value = values[card.Value];
                if ( !value ) {
                    value = card.Value;
                }
                var src = "/nicubunu_White_deck_" + value + "_of_" + cardChars[card.Suit].suit + ".svg";
                img = $('<img src="http://openclipart.org/people/nicubunu' + src + '" />');
                $('#show').append(img);
                img.draggable({start: function(event, ui) { ui.dragable.css('position', 'absolute').css('z-index', 9999) ; } });
            }
        });

    }
}
