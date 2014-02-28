// Copyright (c) 2013 ICRL

// See the file license.txt for copying permission.

var tables = [];
var cSelect = "#b4c8dc";
var cSelectRGB = "rgb(180, 200, 220)"
var cUSelect = "white";

var units = ["Byte", "KB", "MB", "GB", "TB",  "PB", "EB", "ZB"];

var getElements = function(table, destination) {
    table.elements = [];    

    $.ajax({
	async: true,
	// cache: false,
	dataType: 'json',
	type: 'GET',
	url: table.elementUrl,
	data:{ path: table.path, dest: destination},
	success: function(data) {	    
	    table.elements = data;
	    fillTable(table);	    
	    table.path = data[0].path
	    table.pathInformation.innerHTML = table.path
	}
    });

    readLog()
}

var fillTable = function(table) {

    table.tabledom.html("");
    $.each(table.elements, function(i, n){

	if (n.ext == "DUMMY0") {
	    return
	}
	
	var domTable = document.getElementById(table.tableId.replace("#", ""));
	var row = domTable.insertRow(i);

	var icon = row.insertCell(0)
	var name = row.insertCell(1);
	var ext = row.insertCell(2);
	var size = row.insertCell(3);
	var date = row.insertCell(4);
	
	icon.innerHTML = "<div class='ext" + n.ext.replace("\.", "_") + "'></div>"
	name.innerHTML =  n.name;
	ext.innerHTML = n.ext
	size.innerHTML = n.size;
	date.innerHTML = n.date;
    });

    addTableEvents(table);    
}

var addTableEvents = function(table) {
    $(table.tableId + "> tr").each(function (i, n) {
	// this.onclick = function (row) {
	//     rowSelect(this, cPressed, table);
	// }

	this.onclick = function() {
	    $(table.tableId + "> tr").each(function (i_, n_) {
		n_.style.background = cUSelect
	    });

	    n.style.background = cSelect
	}

	this.ondblclick = function () {
	    cd(n, table);
	}

    });
}

var cd = function (row, table) {
    var ext = row.cells[2].innerHTML;

    if (ext != "DIR") {
	return
    }
    
    getElements(table, row.cells[1].innerHTML);
    fillTable(table);    
}

var restore = function() {
    element = ""
    $(tables[0].tableId + "> tr").each(function (i, n) {	
	if (n.style.background.indexOf(cSelectRGB) != -1) {
	    element = n.cells[1].innerHTML
	    return
	}
    })
    
    $.ajax({
	async: true, 	
	type: 'GET',
	url: tables[0].restoreUrl,
	data: {
	    source: tables[0].path,
	    element: element,
	    target: tables[1].path
	},
	success: function() {
	    // trick to reload: just send empty path and path as destination
	    
	    path_ = tables[1].path
	    tables[1].path = ""	    
	    getElements(tables[1], path_)
	}
    });

    readLog()
}

var swapSelection = function() {
    tables[0].selected = !tables[0].selected
    tables[1].selected = !tables[1].selected

    $(".filetree > div:nth-child(1)").css("border", "0px solid black")
    $(".filetree > div:nth-child(2)").css("border", "0px solid black")

    if (tables[0].selected) {
	$(".filetree > div:nth-child(1)").css("border", "1px solid blue")
	return
    }

    $(".filetree > div:nth-child(2)").css("border", "1px solid blue")
}

var selectNext = function(shift, direction) {    
    table = tables[0]

    if (tables[1].selected) {
	table = tables[1]
    }

    $(table.upId).css("background", "")

    index = -1
    $(table.tableId + "> tr").each(function (i, n) {
	if (n.style.background.indexOf(cSelectRGB) != -1) {
	    index = i
	}
	if (!shift) {
	    n.style.background = ""
	}
    });

    if (index+direction < 0) {
	$(table.upId).css("background", cSelect)
	return
    }

    $(table.tableId + " > tr").each(function (i, n) {
	if (i == index+direction) {
	    n.style.background = cSelect
	}
    })
}

var readLog = function() {
    $.ajax({
	type: 'GET',
	async: true,
	url: 'web.log',
	dataType: 'text',
	success: function(data) {
	    var lines = data.split("\n")
	    
	    
	    // data = data.replace(/\r?\n/g, "<br />");
	    $("#log").html("<a href='web.log'>(" + (lines.length - 1) + ") Error Messages </a>")
	}
    });
}

$(document).ajaxStart(function(){
    $('#loadimage').show();
});

$(document).ajaxStop(function(){
    $('#loadimage').hide();    
});

$(document).ready(function (){

    tables = [    
	{
	    id:                "rdiff",
	    elements:          [],
	    
	    elementUrl:        'rdiff-elements',
	    restoreUrl:        'restore',	    

	    tabledom:          $("#rdiff-table"),
	    tableId:           "#rdiff-table",	    
	    fileTable:         $("#rdiff-table"),	    
	    pathInformation:   $(".path-information")[0],
	    path:              "",
	    upId:              "#rdiff-up",
	    selected:          true
	},
	{
	    id:                "physical",
	    elements:          [],	    	    
	    
	    repoUrl:           'physicalrepo',
	    elementUrl:        'phy-elements',	    

	    tabledom:          $("#physical-table"),
	    tableId:           "#physical-table",	    
	    fileTable:         $("#physical-table"),	    
	    pathInformation:   $(".path-information")[1],
	    path:              "",
	    upId:              "#phy-up",
	    selected:          false
	}
    ];    

    document.body.onkeydown = function(key) {

	if (key.keyCode == 83) {
	    swapSelection()
	}

	if (key.keyCode == 40) {
	    key.preventDefault()	    
	    selectNext(key.shiftKey, 1)
	}

	if (key.keyCode == 38) {
	    key.preventDefault()
	    selectNext(key.shiftKey, -1)
	}

	if (key.keyCode == 13) {
	    table = tables[0]
	    if (tables[1].selected) {
		table = tables[1]
	    }

	    upSelect = $(table.upId).css("background")
	    if (upSelect.indexOf(cSelectRGB) != -1) {
		getElements(table, "")
		return
	    }
	    
	    $(table.tableId + "> tr").each(function (i, n) {		
		if (n.style.background == cSelectRGB) {
		    cd(n, table)
		}
	    })
	}

	// arrow down = 40
	// arrow up = 38
	// enter = 13
	
    }
    $(tables[0].upId).bind('click', function() {
	//just get elements with empty target (server jbows what to do)
	getElements(tables[0], "")
    });
    $(tables[1].upId).bind('click', function() {
	//just get elements with empty target (server jbows what to do)
	getElements(tables[1], "")
    });
    $("#restore").bind('click', function() {
	restore()
    })

    getElements(tables[0], "", true);
    getElements(tables[1], "", true, 0);

    readLog()
});
