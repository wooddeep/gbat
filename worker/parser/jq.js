var singleValueOperators = {$gt:1, $gte:1, $lt:1, $lte:1, $not:1} // note that $not is only this type if it has no 'parts'
var possibleMultiValueOperators = {$eq:1, $ne:1}
var arrayOperators = {$nin:1, $all:1, $in:1}
//var specialSingleValueOperators = {$geoIntersects:1, $geoWithin:1} // treating as noValueOperators
//var noValueIndependentOperators = {$where:1, $comment:1}
//var noValueFieldOperators = {$mod:1, $exists:1, $regex:1, $size:1, $nearSphere:1, $near:1}


// builds an object immediate where keys can be expressions

var simpleComparators = {
    $eq: mongoEqual,
    $gt: function(a,b) {return a>b},
    $gte:function(a,b) {return a>=b},
    $lt:function(a,b) {return a<b},
    $lte:function(a,b) {return a<=b},
    $ne:function(a,b) {return a!==b},

    $mod:function(docValue,operand) {return docValue%operand[0] === operand[1]},
    $regex:function(docValue,operand) {return typeof(docValue) === 'string' && docValue.match(RegExp(operand)) !== null},

    $exists:function(docValue,operand) {return (docValue !== undefined) === operand},

    $in:function(docVal,operand) {
        if(Array.isArray(docVal)) {
            return docVal.some(function(val) {
                return operand.indexOf(val) !== -1;
            });
        } else {
            return operand.indexOf(docVal) !== -1
        }
    },
    $nin:function(docVal,operand) {
        if(Array.isArray(docVal)) {
            return docVal.every(function(val) {
                return operand.indexOf(val) === -1;
            });
        } else {
            return operand.indexOf(docVal) === -1
        }
    },
    $all:function(docVal,operand) {
        return docVal instanceof Array && docVal.reduce(function(last,cur) {
                return last && operand.indexOf(cur) !== -1
            },true)
    },
}

var compoundOperatorComparators = {
    $and: function(document, parts) {
        for(var n=0;  n<parts.length; n++) {
            if(!matches(parts[n].parts, document)) {
                return false
            }
        }
        // else
        return true
    },
    $or: function(document, parts) {
        for(var n=0;  n<parts.length; n++) {
            if(matches(parts[n].parts, document)) {
                return true
            }
        }
        // else
        return false
    },
    $nor: function(document, parts) {
        for(var n=0;  n<parts.length; n++) {
            if(matches(parts[n].parts, document)) {
                return false
            }
        }
        // else
        return true
    }
}

var matches = function(parts, document, validate) {
    if(validate !== false)
        validateDocumentObject(document)

    return parts.every(function(part) { return partMatches(part, document, validate)});
}

function partMatches(part, document, validate) {
    var pointers = DotNotationPointers(document, part.field)
    for(var p=0; p<pointers.length; p++) {
        var pointer = pointers[p]

        if(part.operator in simpleComparators) {
            var test = valueTest(pointer.val, part.operand, simpleComparators[part.operator])
            if(!test)
                continue; // this part doesn't match
        } else if(part.operator in compoundOperatorComparators) {
            if(!compoundOperatorComparators[part.operator](document, part.parts)) {
                continue; // this part doesn't match
            }
        } else if(part.operator === '$not') {
            if(part.parts.length > 0) {
                if(matches(part.parts, document, validate)) {
                    continue; // this part doesn't match
                }
            } else {
                if(valueTest(pointer.val, part.operand, mongoEqual) === true)
                    continue; // this part doesn't match
            }
        } else if(part.operator === '$size') {
            return pointer.val instanceof Array && pointer.val.length === part.operand

        } else if(part.operator === '$elemMatch') {
            var documentField = pointer.val
            if(documentField === undefined)
                continue; // this part doesn't match

            if(part.implicitField) {
                for(var n=0; n<part.parts.length; n++) {
                    part.parts[n].field = 'x' // some fake field so it can be tested against
                }
            }

            var anyMatched = false
            for(var n=0; n<documentField.length; n++) {
                if(part.implicitField) {
                    var documentToMatch = {x:documentField[n]}
                } else {
                    var documentToMatch = documentField[n]
                }

                if(matches(part.parts, documentToMatch, validate)) {
                    anyMatched = true
                    break;
                }
            }
            if(!anyMatched)
                continue; // this part doesn't match

        } else if(part.operator === '$where') {
            if(part.field !== undefined) {
                var objectContext = pointer.val
            } else {
                var objectContext = document
            }

            if(!part.operand.call(objectContext))
                continue; // this part doesn't match
        } else if(part.operator === '$comment') {
            return true // ignore it

        } else {
            throw new Error("Unsupported operator: "+part.operator)
        }
        // else
        return true
    }
    // else
    return false
}


// tests a document value against a query value, using a comparison function
// this handles array-contains behavior
function valueTest(documentValue, queryOperand, compare) {
    if(documentValue instanceof Array) {
        if(queryOperand instanceof Array) {
            if(!compare(documentValue, queryOperand)) return false
        } else {
            var test = documentValue.reduce(function(last, cur) {
                return last || compare(cur, queryOperand)
            },false)

            if(!test) return false
        }
    } else {
        return compare(documentValue, queryOperand)
    }
    // else
    return true
}

// matches any value, with mongo's special brand of very strict object equality and weird null matching
function mongoEqual(documentValue,queryOperand) {
    if(documentValue instanceof Array) {
        if(!(queryOperand instanceof Array))
            return false
        if(documentValue.length !== queryOperand.length) {
            return false
        } else {
            return documentValue.reduce(function(previousValue, currentValue, index) {
                return previousValue && mongoEqual(currentValue,queryOperand[index])
            }, true)
        }
    } else if(documentValue instanceof Object) {
        if(!(queryOperand instanceof Object))
            return false

        var aKeys = Object.keys(documentValue)
        var bKeys = Object.keys(queryOperand)

        if(aKeys.length !== bKeys.length) {
            return false
        } else {
            for(var n=0; n<aKeys.length; n++) {
                if(aKeys[n] !== bKeys[n]) return false

                var key = aKeys[n]
                var aVal = documentValue[key]
                var bVal = queryOperand[key]

                if(!mongoEqual(aVal,bVal)) {
                    return false
                }
            }
            // else
            return true
        }
    } else {
        if(queryOperand === null) {
            return documentValue === undefined || documentValue === null
        } else {
            return documentValue===queryOperand
        }
    }
}

function validateDocumentObject(document) {

    Object.keys(document).forEach(function(key) {
        if(key[0] === '$')
            throw new Error("Field names can't start with $")
        else if(key.indexOf('.') !== -1)
            throw new Error("Field names can't contain .")
        else if(document[key] instanceof Object) {
            validateDocumentObject(document[key])
        }
    });

}

var DotNotationPointers = function(rootObject, property) {
    if(property === undefined) {
        property = []
    } else if(!(property instanceof Array)) {
        property = property.split('.')
    }

    return createPointers(rootObject, property)
}

function createPointers(rootObject, propertyParts) {
    var initialObject = {dummy: rootObject}
    var curInfoObjects = [{obj: initialObject, last: 'dummy', propertyPath: []}]

    propertyParts.forEach(function(part) {
        var nextInfoObjects = []
        curInfoObjects.forEach(function(current) {
            var curValue = getValue(current.obj, current.last)
            if(curValue instanceof Array && !isInteger(part)) {
                curValue.forEach(function(property, index) {
                    nextInfoObjects.push({obj: getValue(curValue, index), propertyPath: current.propertyPath.concat(index, part), last: part})
                })
            } else {
                nextInfoObjects.push({obj: curValue, propertyPath: current.propertyPath.concat(part), last: part})
            }
        })

        curInfoObjects = nextInfoObjects
    })

    return curInfoObjects.map(function(current) {
        if(current.obj === initialObject) {
            var obj = current.obj.dummy
            var last = undefined
        } else {
            var obj = current.obj
            var last = current.last
        }
        return new DotNotationPointer(rootObject, current.propertyPath, {obj:obj, last: last})
    })
}

function getValue(object, key) {
    if(object === undefined)
        return undefined
    else
        return object[key]
}

var DotNotationPointer = function(rootObject, property, propertyInfo) {
    this.root = rootObject
    if(property === undefined) {
        this.property = []
    } else if(property instanceof Array) {
        this.property = property
    } else {
        this.property = property.split('.')
    }

    if(propertyInfo !== undefined) {
        this.propertyInfo = propertyInfo
    }
}
DotNotationPointer.prototype = {}

// getter and setter for the value being pointed to
Object.defineProperty(DotNotationPointer.prototype, 'val', {
    get: function() {
        var info = this.propertyInfo
        if(info.obj === undefined) {
            return undefined
        } else {
            if(info.last !== undefined) {
                return info.obj[info.last]
            } else {
                return info.obj
            }
        }
    }, set: function(value) {
        if (value === undefined) {
            if (this.propertyInfo.obj !== undefined) {
                delete this.propertyInfo.obj[this.propertyInfo.last]
            }
        } else {
            if(this.propertyInfo.obj === undefined) { // create the path if it doesn't exist
                createProperty(this)
            }

            this.propertyInfo.obj[this.propertyInfo.last] = value
        }
    }
})


function createProperty(that) {
    var result = that.root
    var lastIndex = that.property.length-1
    for(var n=0; n<lastIndex; n++) {
        var value = result[that.property[n]]
        if(value === undefined) {
            if(isInteger(that.property[n+1]))
                var newValue = []
            else
                var newValue = {}

            value = result[that.property[n]] = newValue
        }

        result = value
    }

    that.propertyInfo = {obj:result, last: that.property[lastIndex]}
}


function isInteger(v) {
    var number = parseInt(v)
    return !isNaN(number)
}

var Parse = function(mongoQuery) {
    this.parts = parseQuery(mongoQuery)
}
Parse.prototype = {}

Parse.prototype.matches = function(document, validate) {
    return matches(this.parts, document, validate)
}


var parse = function(mongoQuery) {
    return new Parse(mongoQuery)
}


var complexFieldIndependantOperators = {$and:1, $or:1, $nor:1}
var simpleFieldIndependantOperators = {$text:1, $comment:1}

// compares two documents by a single sort property
function sortCompare(a,b,sortProperty) {
    var aVal = DotNotationPointers(a, sortProperty)[0].val // todo: figure out what mongo does with multiple matching sort properties
    var bVal = DotNotationPointers(b, sortProperty)[0].val

    if(aVal > bVal) {
        return 1
    } else if(aVal < bVal) {
        return -1
    } else {
        return 0
    }
}

function isInclusive(projection) {
    for(var k in projection) {
        if(!projection[k]) {
            if(k !== '_id') {
                return false
            }
        } else if(k === '$meta') {
            return true
        } else if(projection[k]) {
            if(projection[k] instanceof Object && ('$elemMatch' in projection[k] || '$slice'  in projection[k])) {
                // ignore
            } else {
                return true
            }
        }
    }
}

function parseQuery(query) {
    if(query instanceof Function || typeof(query) === 'string') {
        if(query instanceof Function) {
            query = "("+query+").call(obj)"
        }

        var normalizedFunction = new Function("return function(){var obj=this; return "+query+"}")()
        return [new Part(undefined, '$where', normalizedFunction)]
    }
    // else

    var parts = []
    for(var key in query) {
        if(key in complexFieldIndependantOperators) { // a field-independant operator
            var operator = key
            var operands = query[key]
            var innerParts = []
            operands.forEach(function(operand) {
                innerParts.push(new Part(undefined, '$and', [operand], parseQuery(operand)))
            })

            parts.push(new Part(undefined, operator, query[key], innerParts))
        } else if(key in simpleFieldIndependantOperators) {
            parts.push(new Part(undefined, key, query[key]))
        } else { // a field
            var field = key
            if(isObject(query[key]) && fieldOperand(query[key])) {
                for(var innerOperator in query[key]) {
                    var innerOperand = query[key][innerOperator]
                    parts.push(parseFieldOperator(field, innerOperator, innerOperand))
                }
            } else { // just a value, shorthand for $eq
                parts.push(new Part(field, '$eq', query[key]))
            }
        }
    }

    return parts
}

function map(parts, callback) {
    var result = {}
    parts.forEach(function(part) {
        if(part.operator === '$and') {
            var mappedResult = map(part.parts, callback)
        } else if(part.operator in complexFieldIndependantOperators) {
            var mappedParts = part.parts.map(function(part) {
                return map(part.parts, callback)
            })
            var mappedResult = {$or: mappedParts}
        } else {
            var value = {}; value[part.operator] = part.operand
            var cbResult = callback(part.field, value)
            var mappedResult = processMappedResult(part, cbResult)
        }

        mergeQueries(result, mappedResult)
    })

    compressQuery(result)
    return result

    function processMappedResult(part, mappedResult) {
        if(mappedResult === undefined) {
            var result = {}
            if(part.field === undefined) {
                result[part.operator] = part.operand
            } else {
                var operation = {}
                operation[part.operator] = part.operand
                result[part.field] = operation
            }

            return result
        } else if(mappedResult ===  null) {
            return {}
        } else {
            return mappedResult
        }
    }
}

// merges query b into query a, resolving conflicts by using $and (or other techniques)
function mergeQueries(a,b) {
    for(var k in b) {
        if(k in a) {
            if(k === '$and') {
                a[k] = a[k].concat(b[k])
            } else {
                var andOperandA = {}; andOperandA[k] = a[k]
                var andOperandB = {}; andOperandB[k] = b[k]
                var and = {$and:[andOperandA,andOperandB]}
                delete a[k]
                mergeQueries(a,and)
            }
        } else {
            a[k] = b[k]
        }
    }
}

// decanonicalizes the query to remove any $and or $eq that can be merged up with its parent object
// compresses in place (mutates)
var compressQuery = function (x) {
    for(var operator in complexFieldIndependantOperators) {
        if(operator in x) {
            x[operator].forEach(function(query){
                compressQuery(query)
            })
        }
    }
    if('$and' in x) {
        x.$and.forEach(function(andOperand) {
            for(var k in andOperand) {
                if(k in x) {
                    if(!(x[k] instanceof Array) && typeof(x[k]) === 'object' && k[0] !== '$') {
                        for(var operator in andOperand[k]) {
                            if(!(operator in x[k])) {
                                x[k][operator] = andOperand[k][operator]
                                delete andOperand[k][operator]
                                if(Object.keys(andOperand[k]).length === 0)
                                    delete andOperand[k]
                            }
                        }
                    }
                } else {
                    x[k] = andOperand[k]
                    delete andOperand[k]
                }
            }
        })
        x.$and = filterEmpties(x.$and)
        if(x.$and.length === 0) {
            delete x.$and
        }
    }
    if('$or' in x) {
        x.$or = filterEmpties(x.$or)
        if(x.$or.length === 0) {
            delete x.$or
        } else if(x.$or.length === 1) {
            var orOperand = x.$or[0]
            delete x.$or
            mergeQueries(x,orOperand)
        }
    }

    for(var k in x) {
        if(x[k].$eq !== undefined && Object.keys(x[k]).length === 1) {
            x[k] = x[k].$eq
        }
        if(x[k].$elemMatch !== undefined) {
            compressQuery(x[k].$elemMatch)
        }
    }

    return x

    function filterEmpties(a) {
        return a.filter(function(operand) {
            if(Object.keys(operand).length === 0)
                return false
            else
                return true
        })
    }
}

// returns a Part object
function parseFieldOperator(field, operator, operand) {
    if(operator === '$elemMatch') {
        var elemMatchInfo = parseElemMatch(operand)
        var innerParts = elemMatchInfo.parts
        var implicitField = elemMatchInfo.implicitField
    } else if(operator === '$not') {
        var innerParts = parseNot(field, operand)
    } else {
        var innerParts = []
    }
    return new Part(field, operator, operand, innerParts, implicitField)
}

// takes in the operand of the $elemMatch operator
// returns the parts that operand parses to, and the implicitField value
function parseElemMatch(operand) {
    if(fieldOperand(operand)) {
        var parts = []
        for(var operator in operand) {
            var innerOperand = operand[operator]
            parts.push(parseFieldOperator(undefined, operator, innerOperand))
        }
        return {parts: parts, implicitField: true}
    } else {          // non-field operators ( stuff like {a:5} or {$and:[...]} )
        return {parts: parseQuery(operand), implicitField: false}
    }
}

function parseNot(field, operand) {
    var parts = []
    for(var operator in operand) {
        var subOperand = operand[operator]
        parts.push(parseFieldOperator(field, operator, subOperand))
    }
    return parts
}

// returns true for objects like {$gt:5}, {$elemMatch:{...}}
// returns false for objects like {x:4} and {$or:[...]}
function fieldOperand(obj) {
    for(var key in obj) {
        return key[0] === '$' && !(key in complexFieldIndependantOperators) // yes i know this won't actually loop
    }
}

// returns true if the value is an object and *not* an array
function isObject(value) {
    return value instanceof Object && !(value instanceof Array)
}

var Part = function(field, operator, operand, parts, implicitField) {
    if(parts === undefined) parts = []

    this.field = field
    this.operator = operator
    this.operand = operand
    this.parts = parts
    this.implicitField = implicitField // only used for a certain type of $elemMatch part
}


var newQuery = {$and:[{userId: 1}, {animal: {$in: ['beefalo', 'deerclops']}}]}
var ret = parse(newQuery).matches({userId: 1, animal: 'deerclops'}) // returns true
console.log(ret)

