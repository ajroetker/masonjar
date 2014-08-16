var suits = {
    1 : 'spades',
    2 : 'hearts',
    3 : 'clubs',
    4 : 'diamonds',
}
var values = {
    1  : 'Ace',
    2  : '2',
    3  : '3',
    4  : '4',
    5  : '5',
    6  : '6',
    7  : '7',
    8  : '8',
    9  : '9',
    10 : '10',
    11 : 'Jack',
    12 : 'Queen',
    13 : 'King'
}

var suitsRev = {
    'spades'   : 1,
    'hearts'   : 2,
    'clubs'    : 3,
    'diamonds' : 4,
}
var valuesRev = {
    'Ace'   : 1,
    '2'     : 2,
    '3'     : 3,
    '4'     : 4,
    '5'     : 5,
    '6'     : 6,
    '7'     : 7,
    '8'     : 8,
    '9'     : 9,
    '10'    : 10,
    'Jack'  : 11,
    'Queen' : 12,
    'King'  : 13,
}

var idParserMap = {
    'd' : 'diamonds',
    'h' : 'hearts',
    's' : 'spades',
    'c' : 'clubs',
}

function cardifyId( id ){
    var suitStr = idParserMap[ id[0] ];
    //Doesn't quite work
    var valueStr = id.slice( suitStr.length, id.length );
    var card = {
        Suit  : suitsRev[ suitStr ],
        Value : valuesRev[ valueStr ],
        //Figure out a good way to do this
        Owner : 'test@example.com',
    }
    return card
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

function padding( element ){
    return parseFloat( element.css('padding') );
}

function height( element ){
    return parseFloat( element.height() );
}

function width( element ){
    return parseFloat( element.width() );
}

function lastElement( array ){
    return array[ array.length - 1 ];
}

function Board(object) {
    this.nertz  = object.Nertz;
    this.renderNertz = function() {
        var nertzPileLength = this.nertz.length;
        $.each( this.nertz, function( index, card ) {
            if ( card.Value ) {
                var topPos = index * height( $( '.card' ) ) * 0.03 ;
                if (index < nertzPileLength - 1) {
                    img = cardImgBack();
                    $('#nertz').append( img );
                    if ( index === 0 ) {
                      topPos = padding( $('.pile') );
                    };
                    img.css( 'top', topPos )
                       .data('pile', 'nertz');
                } else {
                    img = cardImg( suits[card.Suit], values[card.Value] );
                    $('#nertz').append(img);
                    img.css( 'top', topPos )
                       .data('pile', 'nertz');
                    img.draggable({
                        revert: 'invalid',
                        zIndex: 100,
                    });
                };
            }
        });
    };
    this.stream = object.Stream;
    this.river  = object.River;
    this.renderRiver = function() {
        var board = this;
        $.each( this.river, function( jndex, cards ) {
            if (cards.length === 0) {
                // Replace this with an arbitrary droppable
                board.river[jndex].push( board.nertz.pop() );
            };
            $.each( cards, function( index, card ) {
                if ( card.Value ) {
                    img = cardImg( suits[card.Suit], values[card.Value] );
                    $( '#river' + jndex ).append(img);
                    var topPos = index * height( $( '.card' ) ) * 0.25 ;
                    if ( index === 0 ) {
                      topPos = padding( $('.pile') );
                    };
                    img.css( 'top', topPos)
                       .data('pile', 'river')
                       .data('river', jndex)
                       .data('index', index);
                    if ( index === cards.length - 1 ) {
                        img.droppable({
                            accept: function(dropped) {
                                return dropped.attr('id') === ( suits[( card.Suit % 2 ) + 1] + values[card.Value - 1] ) ||
                                       dropped.attr('id') === ( suits[( card.Suit % 2 ) + 3] + values[card.Value - 1] ) ;
                            },
                            drop: function(event, ui) {
                                $( event.target ).droppable('disable');
                                var toElement = $( event.toElement );
                                var pile = toElement.data('pile');
                                var droppedCard;
                                switch ( pile ) {
                                    case 'river':
                                        var riverPile = board.river[ parseInt(toElement.data('river')) ];
                                        var subPileIndex = parseInt(toElement.data('index'));
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
                        // Arbitrary to be on top
                        zIndex: 100,
                        helper: function() {
                            var wholePile = $( '#river' + jndex + ' img' ).get();
                            var w = width( $( '#river' + jndex ) );
                            var slice = wholePile.slice(index, wholePile.length);
                            var h = ( padding( $('.pile') ) * 2 )
                                    + height( $( '.card' ) )
                                    + ( ( slice.length - 1 ) * height( $( '.card' ) ) * 0.25 );
                            var container =
                                $('<div/>').attr('id', 'draggingContainer')
                                           .width( w ).height( h );
                            $.each( slice, function( kndex, card ) {
                                var topPos = kndex * height( $( '.card' ) ) * 0.25 ;
                                if ( kndex === 0 ) {
                                    topPos = padding( $('.pile') );
                                };
                                $( card ).clone()
                                         .css( 'top', topPos )
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
        var topPadding = padding( $('.pile') );
        this.renderRiver();
        this.renderNertz();
        var streamPileLength = this.stream.length;
        var board = this;
        if ( this.stream[0] ) {
            $.each( this.stream, function( index, card ) {
                if ( card.Value ) {
                    img = cardImgBack();
                    $('#stream').append(img);
                    img.css( 'top',  topPadding + index + "%" )
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
                img = cardImg( suits[card.Suit], values[card.Value] )
                $('#show').append(img);
                img.css( 'top',  topPadding + index + "%" )
                   .data('pile', 'show');
                img.draggable({
                    zIndex: 100,
                    revert: 'invalid',
                });
            }
        });

    }
}


