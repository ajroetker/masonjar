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
    return $('<img id="' + suit + '' + value +
             '" class="card" src="http://openclipart.org/people/' +
             'nicubunu/nicubunu_White_deck_' +
             value + '_of_' + suit + '.svg" />');
}

function cardImgBack() {
    return $('<img class="card" src="http://openclipart.org/people/' +
             'nicubunu/nicubunu_Card_backs_simple_blue.svg" />');
}

function Board(object) {
    this.nertz  = object.Nertz;
    this.renderNertz = function( globalPosition ) {
        var nertzPileLength = this.nertz.length;
        $.each( this.nertz, function( index, card ) {
            if ( card.Value ) {
                if (index < nertzPileLength - 1) {
                    img = cardImgBack();
                    $('#nertz').append( img );
                    var topsPos;
                    if (index) {
                      topPos = index * parseFloat(img.height()) * 0.03 ;
                    } else {
                      topPos = $('#nertz').css('padding');
                    };
                    img.css( 'top', topPos )
                       .data('pile', 'nertz');
                } else {
                    img = cardImg( cardChars[card.Suit].suit, values[card.Value] );
                    $('#nertz').append(img);
                    var topPos = index * parseFloat( $('.card').height()) * 0.03 ;
                    img.css( 'top', topPos )
                       .data('pile', 'nertz');
                    img.draggable({
                        revert: 'invalid',
                        stack: '#boardGame div',
                    });
                };
            }
        });
    };
    this.stream = object.Stream;
    this.river  = object.River;
    this.renderRiver = function( globalPosition ) {
        var board = this;
        $.each( this.river, function( jndex, cards ) {
            if (cards.length === 0) {
                board.river[jndex].push( board.nertz.pop() );
            };
            $.each( cards, function( index, card ) {
                if ( card.Value ) {
                    img = cardImg( cardChars[card.Suit].suit, values[card.Value] );
                    $('#river' + jndex).append(img);
                    var topsPos;
                    if (index) {
                      topPos = index * parseFloat(img.height()) * 0.25 ;
                    } else {
                      topPos = $('#river' + jndex ).css('padding');
                    };
                    img.css( 'top', topPos )
                       .data('pile', 'river')
                       .data('river', jndex)
                       .data('index', index);
                    if ( index === cards.length - 1 ) {
                        img.droppable({
                            accept: function(dropped) {
                                return dropped.attr('id') === ( cardChars[( card.Suit % 2 ) + 1].suit + values[card.Value - 1] ) ||
                                       dropped.attr('id') === ( cardChars[( card.Suit % 2 ) + 3].suit + values[card.Value - 1] ) ;
                            },
                            drop: function(event, ui) {
                                $( event.target ).droppable('disable');
                                var pile = $( event.toElement ).data('pile');
                                var droppedCard;
                                switch ( pile ) {
                                    case 'river':
                                        var riverPileNum = parseInt($( event.toElement ).data('river'));
                                        var riverPile = board.river[riverPileNum];
                                        var subPileIndex = parseInt($( event.toElement ).data('index'));
                                        var subPile = riverPile.slice( subPileIndex, riverPile.length );
                                        subPile.reverse();
                                        while (subPile.length > 0){
                                            var popped = subPile.pop();
                                            cards.push( popped );
                                        }
                                        riverPile.splice( subPileIndex, riverPile.length - subPileIndex );
                                        break;
                                    case 'nertz':
                                        droppedCard = board.nertz.pop();
                                        cards.push( droppedCard );
                                        break;
                                    case 'show':
                                        droppedCard = board.show.pop();
                                        cards.push( droppedCard );
                                        break;
                                }
                                board.render();
                            }
                        });
                    };
                    img.draggable({
                        revert: 'invalid',
                        //cursorAt: { left: 5, top: 5 },
                        // Arbitrary to be on top
                        zIndex: 100,
                        helper: function() {
                            var wholePile = $( '#river' + jndex + ' img' ).get();
                            var w = $( '#river' + jndex ).width();
                            var slice = wholePile.slice(index, wholePile.length);
                            var h = ( parseInt( $( '#river' + jndex ).css('padding') ) * 2 )
                                    + parseFloat( $( slice[0] ).height() )
                                    + ( ( slice.length - 1 ) * parseFloat( $( slice[0] ).height() ) * 0.25 );
                            var container =
                                $('<div/>').attr('id', 'draggingContainer')
                                           .width( w ).height( h );
                            $.each( slice, function( kndex, card ) {
                                var tmp =
                                $( card ).clone()
                                         .addClass( 'card' )
                                         .data( 'pile', 'river' )
                                         .data( 'river', jndex )
                                         .data( 'index', kndex )
                                         .appendTo( container );
                            });
                            $('#draggingContainer').css( 'bottom', $('#draggingContainer:last-child').css('bottom') );
                            return container;
                        },
                    });
                }
            });
        });
    };
    this.show   = object.Show;
    this.render = function() {
        $('#boardGame div img').remove();
        var t = parseFloat( $('.pile').css( 'padding' ) );
        var gps = { top : t };
        this.renderRiver(gps);
        this.renderNertz(gps);
        var streamPileLength = this.stream.length;
        var board = this;
        if ( this.stream[0] ) {
            $.each( this.stream, function( index, card ) {
                if ( card.Value ) {
                    img = cardImgBack();
                    $('#stream').append(img);
                    img.css( 'top',  gps.top + ( index * 1 ) + "%" )
                       .data('pile', 'stream');
                    if ( index === streamPileLength - 1 ) {
                        img.on( "click", function() {
                            if ( board.stream[0] ) { board.show.push( board.stream.pop() ) };
                            if ( board.stream[0] ) { board.show.push( board.stream.pop() ) };
                            if ( board.stream[0] ) { board.show.push( board.stream.pop() ) };
                            board.render();
                        });
                    };
                }
            });
        } else {
            $('#stream').on( "click", function() {
                while ( board.show.length > 0 ) {
                    board.stream.push( board.show.pop() );
                };
                board.render();
                $('#stream').off( "click" );
            });
        };
        $.each( this.show, function( index, card ) {
            if ( card.Value ) {
                img = cardImg( cardChars[card.Suit].suit, values[card.Value] )
                $('#show').append(img);
                img.css( 'top',  gps.top + ( index * 1 ) + "%" )
                   .data('pile', 'show');
                img.draggable({
                    stack: '#boardGame div',
                    revert: 'invalid',
                });
            }
        });

    }
}


