MathJax = {
	output: {
		scale: 1, // global scaling factor for all expressions
		minScale: 0.4, // smallest scaling factor to use
		mtextInheritFont: false, // true to make mtext elements use surrounding font
		merrorInheritFont: false, // true to make merror text use surrounding font
		mtextFont: "", // font to use for mtext, if not inheriting (empty means use MathJax fonts)
		merrorFont: "", // font to use for merror, if not inheriting (empty means use MathJax fonts)
		unknownFamily: "", // font to use for character that aren't in MathJax's fonts
		mathmlSpacing: false, // true for MathML spacing rules, false for TeX rules
		skipAttributes: {}, // RFDa and other attributes NOT to copy to the output
		exFactor: 0.5, // default size of ex in em units
		displayAlign: "auto", // default for indentalign when set to 'auto'
		displayIndent: "0", // default for indentshift when set to 'auto'
		displayOverflow: "scale", // default for overflow (scroll/scale/truncate/elide/linebreak/overflow)
		linebreaks: {
			// options for when overflow is linebreak
			inline: true, // true for browser-based breaking of inline equations
			width: "95%", // a fixed size or a percentage of the container width
			lineleading: 0.1, // the default lineleading in em units
		},
		htmlHDW: "use", // 'use', 'force', or 'ignore' data-mjx-hdw attributes
		preFilters: [], // A list of pre-filters to add to the output jax
		postFilters: [], // A list of post-filters to add to the output jax
	},
};

document.addEventListener("DOMContentLoaded", function (event)
{
	var mob = 1;
	if (window.innerWidth < 450)
	{
		mob = 0.4;
	}
	if (window.innerWidth < 1300)
	{
		mob = 0.6;
	}
	const INTEGER_SCALING = true;


	const Im3 = 0;
	const Im2 = 1;
	const Im1 = 2;
	const In0 = 3;
	const Il1 = 4;
	const Ic2 = 5;
	const Ie3 = 6;
	// const ST_TOP = 7;
	const ST_ZERO = 3;

	const STINDICES = [
		"is a property expressed by",
		"is contained by",
		"comes from",
		"is near/smimilar to",
		"leads to",
		"contains",
		"expresses property",
	];

	const POST_METHOD = "POST";

	const HISTORY_KEY = "sst-search-history";
	const MAX_HISTORY_ITEMS = 30;

	function saveSearchToHistory(query)
	{
		if (!query) return;

		let history = JSON.parse(localStorage.getItem(HISTORY_KEY)) || [];

		history = history.filter((item) => item !== query);

		history.unshift(query);

		if (history.length > MAX_HISTORY_ITEMS)
		{
			history = history.slice(0, MAX_HISTORY_ITEMS);
		}

		localStorage.setItem(HISTORY_KEY, JSON.stringify(history));
	}

	function loadHistoryIntoDatalist()
	{
		const datalist = document.getElementById("search-suggestions");
		const history = JSON.parse(localStorage.getItem(HISTORY_KEY)) || [];

		if (!datalist) return;

		history.forEach((item) =>
		{
			const option = document.createElement("option");
			option.value = item;
			datalist.appendChild(option);
		});
	}
	/***********************************************************/
	// Graphics config
	/***********************************************************/
	var CANVAS = CreateCanvas();
	var CTX = CANVAS.getContext("2d");

	// Adjust coordinates
	let WIDTH = CANVAS.offsetWidth;
	let HEIGHT = CANVAS.offsetHeight;

	CANVAS.width = WIDTH;

	let ORGX = WIDTH / 2;
	let ORGY = HEIGHT / 2;
	let THETA = Math.PI / 9;
	let PHI = Math.PI / 9;
	let SCALE = 0.9;
	let OBS_X = 1;
	let OBS_Y = 0.5;
	let OBS_Z = -1;
	let VP_X = 0;
	let VP_Y = 4;
	let VP_Z = 8;

	/************************************************************/
	// Page renderers
	/************************************************************/
	window.addEventListener("popstate", (event) =>
	{
		console.log("Popped state:", event.state);

		if (event.state && event.state.searchQuery)
		{
			sendLinkSearch(event.state.searchQuery);
		} else
		{
			FetchPage(); // Reload the initial view.
		}
	});

	function RerenderMath()
	{
		if (
			window.MathJax &&
			window.MathJax.startup &&
			window.MathJax.startup.promise
		)
		{
			window.MathJax.startup.promise
				.then(() =>
				{
					window.MathJax.typesetPromise();
				})
				.catch((err) => console.log("MathJax typeset failed: " + err.message));
		}
	}

	function topFunction()
	{
		document.body.scrollTop = 0;
		document.documentElement.scrollTop = 0;
	}

	async function FetchPage()
	{
		let requestURL = "/searchN4L";
		let request = new Request(requestURL);

		try
		{
			RerenderMath();
			let response = await fetch(request);
			startHipnotize();
			let mynote = await response.json();
			DoHeader(mynote);
			DoOrbitPanel(mynote); // Start in orbit
		} catch (error)
		{
			console.error("FetchPage error:", error);
		} finally
		{
			stopHipnotize();
		}
	}

	function AppRouter()
	{
		const urlParams = new URLSearchParams(window.location.search);

		const searchQuery = urlParams.get("search");

		if (searchQuery)
		{
			console.log("Routing to saved search:", searchQuery);
			sendLinkSearch(searchQuery);
		} else
		{
			console.log("Routing to default page.");
			FetchPage();
		}
	}

	function DisplayError(message)
	{
		const main = document.querySelector("main");
		main.innerHTML = `<div class="error-message"> <h5>An Error Occurred!</h5> <p>${message}</p> </div>`;
	}
	/***********************************************************/
	function DoHeader(obj)
	{
		RerenderMath();
		// Clear the main panel here, as it's common to all
		let clearscreen = document.querySelector("main");
		clearscreen.innerHTML = "";

		// Now manage the header
		let header = document.querySelector("header");
		header.innerHTML = "";

		let titlebar = document.createElement("h2");
		titlebar.id = "topmost_page_title";

		let title = "app";

		switch (obj.Response)
		{
		case "Orbits":
			if (obj.Content != null && obj.Content[0] != null)
			{
				title = obj.Content[0].Text;
			} else
			{
				title = obj.Time;
			}
			break;
		case "ConePaths":
			title = "Local cone paths";
			break;
		case "PathSolve":
			title = "Path solutions";
			break;
		case "Sequence":
			if (obj.Content != null && obj.Content[0] != null)
			{
				title = "Story sequences ... " + obj.Content[0].Chapter;
			} else
			{
				title = "Story sequences ... ";
			}
			break;
		case "PageMap":
			title = "Page notes about " + obj.Content.Title;
			break;
		case "TOC":
			title = "Table of contents";
			break;
		case "Arrows":
			title = "Arrow lookup";
			break;
		default:
			title = "SSToryGraph browser";
			break;
		}

		if (title.length < 60 || IsMath(title))
		{
			titlebar.textContent = title;
		} else
		{
			titlebar.textContent = title.slice(0, 60) + "...";
		}

		header.appendChild(titlebar);

		let nowbar = document.createElement("div");
		nowbar.id = "current_time_context";
		nowbar.textContent = obj.Ambient;
		header.appendChild(nowbar);

		let ctxbar = document.createElement("div");
		ctxbar.id = "context_history";
		ctxbar.textContent = obj.Intent;
		header.appendChild(ctxbar);
	}

	/***********************************************************/
	function DoOrbitPanel(obj)
	{
		RerenderMath();
		let section = document.querySelector("main");
		let panel = document.createElement("i");
		panel.id = "main_content_panel";
		section.appendChild(panel);

		// List of events unrelated
		if (obj == null)
		{
			return;
		}

		CANVAS = CreateCanvas();
		DrawGrid(0, 0, 1);

		const separates = 0;

		if (obj.Content.length == 0)
		{
			panel.textContent = "No result";
			return;
		}

		let last_node_event = obj.Content[0];

		for (let node_event of obj.Content)
		{
			ShowNodeEvent(panel, node_event, separates, "all", "", "h3");

			last_node_event = node_event; // don't link up
			PlotGraphics(node_event, last_node_event);
		}
	}

	/***********************************************************/
	function DoEntireConePanel(obj)
	{
		RerenderMath();
		let section = document.querySelector("main");
		let panel = document.createElement("span");
		panel.id = "main_content_panel";
		section.appendChild(panel);

		CANVAS = CreateCanvas();
		DrawGrid(0, 0, 1);

		// Iterate over the cones from different starting nodes
		for (let head_nptr of obj.Content)
		{
			let nclass = head_nptr.NClass;
			let ncptr = head_nptr.NCPtr;

			let card = document.createElement("div");
			card.setAttribute("class", "card-view");
			panel.appendChild(card);

			let item = document.createElement("h2");
			let link = document.createElement("a");
			item.textContent = head_nptr.Title.slice(0, 50) + "..";

			link.onclick = function ()
			{
				sendlinkData(nclass, ncptr);
			};

			item.appendChild(link);
			card.appendChild(item);

			card = PrintPaths(card, head_nptr.Paths);

			// Add the centrality box at bottom of page
			let tab = document.createElement("table");
			let row = document.createElement("tr");
			let col1 = document.createElement("td");
			let hd1 = document.createElement("h4");

			hd1.textContent = "Betweenness Centrality Rank";
			col1.appendChild(hd1);

			let lst1 = document.createElement("ol");

			// For paths solve <a|b>
			if (head_nptr.BTWC != null)
			{
				for (let centrality of head_nptr.BTWC)
				{
					let li = document.createElement("li");
					li.textContent = centrality;
					lst1.appendChild(li);
				}

				col1.appendChild(lst1);

				let col2 = document.createElement("td");
				let hd2 = document.createElement("h4");
				hd2.textContent = "Supernode summary";
				col2.appendChild(hd2);
				let lst2 = document.createElement("ol");

				for (let snode of head_nptr.Supernodes)
				{
					let li = document.createElement("li");
					li.textContent = snode;
					lst2.appendChild(li);
				}
				col2.appendChild(lst2);

				row.appendChild(col1);
				row.appendChild(col2);

				tab.appendChild(row);
				card.appendChild(tab);
			}
		}
	}

	/***********************************************************/
	function DoSeqPanel(obj)
	{
		let section = document.querySelector("main");
		let panel = document.createElement("div");
		panel.id = "main_content_panel";
		section.appendChild(panel);

		CANVAS = CreateCanvas();
		DrawGrid(0, 0, 1);

		let counter = 1;
		let laststory = obj.Content[0]; // arbitrary init

		for (let story of obj.Content)
		{
			if (counter == 1)
			{
				let item = document.createElement("h2");
				item.textContent = story.Chap;
				panel.appendChild(item);
				let citem = document.createElement("span");
				citem.textContent = "In the context:   " + story.Context;
				panel.appendChild(citem);
			}

			ShowSequenceItem(panel, story, counter, "fwd", "then", "span");
			PlotGraphics(story, laststory);
			laststory = story; // link up

			counter++;
		}
	}

	/***********************************************************/
	function DoPageMapPanel(obj)
	{
		let section = document.querySelector("main");

		let panel = document.createElement("div");
		panel.id = "main_content_panel";
		section.appendChild(panel);

		CANVAS = CreateCanvas();
		DrawGrid(0, 0, 1);

		// Major header
		let title = document.createElement("h3");
		title.textContent = "* Chapter Notes: ";
		title.id = "chapter_notes_heading";
		panel.appendChild(title);

		panel = PrintNotes(panel, obj.Content.Notes);
	}

	/***********************************************************/
	function DoTOCPanel(obj)
	{
		RerenderMath();
		let section = document.querySelector("main");

		let panel = document.createElement("div");
		panel.id = "main_content_panel";
		section.appendChild(panel);

		let item = document.createElement("h3");
		item.textContent = "Table of contents and contexts";
		panel.appendChild(item);

		CANVAS = CreateCanvas();
		DrawGrid(0, 0, 1);
		let counter = 0;

		for (let chpblk of obj.Content)
		{
			counter++;

			// N. Section/chapter header title
			let chapter_section = document.createElement("div");
			chapter_section.setAttribute("class", "card-view");
			chapter_section.id = "toc-panel";

			chapter_section.style.display = "inline-flex";
			let link = document.createElement("a");
			let item = document.createElement("h3");
			link.onclick = function ()
			{
				sendLinkSearch('\\notes \\chapter "' + chpblk.Chapter + '"');
			};
			item.textContent = counter + ". " + chpblk.Chapter;
			link.appendChild(item);
			chapter_section.appendChild(link);

			Event(chpblk.XYZ.X, chpblk.XYZ.Y, chpblk.XYZ.Z);
			Label(chpblk.XYZ.X, chpblk.XYZ.Y, chpblk.XYZ.Z, chpblk.Chapter, 12, "gray",
			);

			// First do the context groups or ambient parts
			if (chpblk.Context != null)
			{
				for (let ctx of chpblk.Context)
				{
					let link = document.createElement("a");
					link.onclick = function ()
					{
						sendLinkSearch("any \\context " + CtxSplice(ctx.Text));
					};
					link.id = "toc-frag";
					link.textContent = "Context Set/Grouping:: ";

					let sitem = document.createElement("span");
					sitem.textContent = ctx.Text;
					sitem.id = "toc-elem";
					link.appendChild(sitem);

					chapter_section.appendChild(link);

					Concept(ctx.XYZ.X, ctx.XYZ.Y, ctx.XYZ.Z);
					Contains(
						chpblk.XYZ.X,
						chpblk.XYZ.Y,
						chpblk.XYZ.Z,
						ctx.XYZ.X,
						ctx.XYZ.Y,
						ctx.XYZ.Z,
					);
				}
			}

			//Spacer ?
			if (chpblk.Single != null)
			{
				for (let ctx of chpblk.Single)
				{
					let link = document.createElement("a");
					link.onclick = function ()
					{
						sendLinkSearch("any \\context " + CtxSplice(ctx.Text));
					};
					link.textContent = "Intentionally emph:: ";

					let sitem = document.createElement("span");
					sitem.textContent = ctx.Text;
					sitem.id = "toc-single";
					link.appendChild(sitem);

					chapter_section.appendChild(link);

					Thing(ctx.XYZ.X, ctx.XYZ.Y, ctx.XYZ.Z);
					Near(
						chpblk.XYZ.X,
						chpblk.XYZ.Y,
						chpblk.XYZ.Z,
						ctx.XYZ.X,
						ctx.XYZ.Y,
						ctx.XYZ.Z,
					);
				}
			}

			if (chpblk.Common != null)
			{
				for (let ctx of chpblk.Common)
				{
					let link = document.createElement("a");
					let sitem = document.createElement("span");
					link.onclick = function ()
					{
						sendLinkSearch("any \\context " + CtxSplice(ctx.Text));
					};
					link.textContent = "Ambient context:: ";
					link.id = "toc-frag";

					sitem.textContent = ctx.Text;
					sitem.id = "toc-common";

					link.appendChild(sitem);
					chapter_section.appendChild(link);

					Thing(ctx.XYZ.X, ctx.XYZ.Y, ctx.XYZ.Z);
					Near(
						chpblk.XYZ.X,
						chpblk.XYZ.Y,
						chpblk.XYZ.Z,
						ctx.XYZ.X,
						ctx.XYZ.Y,
						ctx.XYZ.Z,
					);
				}
			}

			panel.appendChild(chapter_section);
		}
	}

	/***********************************************************/
	function DoStatsPanel(obj)
	{
		RerenderMath();
		let section = document.querySelector("main");

		let panel = document.createElement("div");
		panel.id = "main_content_panel";
		section.appendChild(panel);

		let title = document.createElement("h3");
		title.textContent = "Progress tracker";
		panel.appendChild(title);

		CANVAS = CreateCanvas();
		DrawGrid(0, 0, 1);

		let counter = 0;
		let lastsection = "xxx";
		let link;
		let item;
		let last;
		let hmpanel;
		let card;

		for (let ls of obj.Content)
		{
			if (ls.Section != lastsection)
			{
				counter++;
				lastsection = ls.Section;
				link = document.createElement("a");
				link.onclick = function ()
				{
					sendLinkSearch("\\notes " + '"' + ls.Section + '"');
				};

				card = document.createElement("div");
				card.setAttribute("class", "card-view");
				panel.appendChild(card);

				item = document.createElement("strong");
				item.textContent = counter + ". " + ls.Section;
				link.appendChild(item);
				card.appendChild(link);

				let sec = document.createElement("p");
				sec.id = "toc-panel";

				let sub1 = document.createElement("i");
				sub1.textContent = "Last viewed at " + ls.Last;
				sec.appendChild(sub1);

				let sub2 = document.createElement("i");
				sub2.textContent = "  total viewing count = " + ls.Freq;
				sub2.id = "statcount";
				sec.appendChild(sub2);

				card.appendChild(sec);

				hmpanel = document.createElement("div");
				card.appendChild(hmpanel);

				Concept(ls.XYZ.X, ls.XYZ.Y, ls.XYZ.Z);
				Label(ls.XYZ.X, ls.XYZ.Y, ls.XYZ.Z, ls.Section, 12, "gray");
			} else
			{
				let nlink = document.createElement("a");
				let nitem = document.createElement("span");
				nitem.id = "heatmap";

				if (ls.NPtr.Class < 0)
				{
					nlink.onclick = function ()
					{
						sendLinkSearch('any \\chapter "' + ls.Section + '"');
					};
					nitem.textContent = "browse";
				} else
				{
					nlink.onclick = function ()
					{
						sendLinkSearch("(" + ls.NPtr.Class + "," + ls.NPtr.CPtr + ")");
					};
					nitem.textContent = "(" + ls.NPtr.Class + "," + ls.NPtr.CPtr + ")";
				}
				nitem.style.color = HeatColour(ls.Freq, ls.Pdelta, 70);
				nitem.style.fontSize = "80%";
				nitem.style.backgroundColor = HeatColour(ls.Freq, ls.Pdelta, 100);
				nitem.style.padding = "5px";
				nlink.appendChild(nitem);
				hmpanel.appendChild(nlink);
			}
		}
	}

	/***********************************************************/
	function DoArrowsPanel(obj)
	{
		let section = document.querySelector("main");
		let panel = document.createElement("span");
		panel.id = "main_content_panel";
		section.appendChild(panel);

		CANVAS = CreateCanvas();
		DrawGrid(0, 0, 1);
		let counter = 0;

		let t = document.createElement("h3");
		t.textContent = "Arrows";
		panel.appendChild(t);

		let matches = obj.Content.length;

		let angle_increment = Math.PI / matches;
		let angle = 0;

		let subpanel = document.createElement("div");
		panel.appendChild(subpanel);

		for (let arrow of obj.Content)
		{
			ArrowPair(angle, arrow.ASTtype, arrow.Long, arrow.InvL);
			angle += angle_increment;

			let a = document.createElement("div");
			a.textContent =
				"ArrowPtr = " + arrow.ArrPtr + " (STtype: " + arrow.ASTtype + ")";
			a.id = "arrow-" + (arrow.ASTtype + ST_ZERO);
			subpanel.appendChild(a);

			let ab = document.createElement("p");
			ab.textContent = "Long name: (" + arrow.Long + ")";
			ab.id = "fwdarr";
			//ab.id = 'arrow-'+(arrow.ASTtype + ST_ZERO);
			ab.style.fontFamily = "Verdana";
			a.appendChild(ab);

			let af = document.createElement("p");
			af.textContent = "Short alias: (" + arrow.Short + ")";
			af.id = "fwdarr";
			//af.id = 'arrow-'+(arrow.ASTtype + ST_ZERO);
			af.style.fontFamily = "Verdana";
			a.appendChild(af);

			// Inverse
			let i = document.createElement("div");
			i.textContent =
				"ArrowPtr = " + arrow.InvPtr + ". (STtype: " + arrow.ISTtype + ")";
			//i.id = 'arrow-'+(arrow.ISTtype + ST_ZERO);
			subpanel.appendChild(i);

			let iaf = document.createElement("p");
			iaf.textContent = "Short: " + arrow.InvS;
			//iaf.id = 'arrow-'+(arrow.ISTtype + ST_ZERO);
			iaf.id = "bwdarr";
			iaf.style.fontFamily = "Verdana";
			i.appendChild(iaf);

			let ib = document.createElement("p");
			ib.textContent = "Long: " + arrow.InvL;
			//ib.id = 'arrow-'+(arrow.ISTtype + ST_ZERO);
			ib.id = "bwdarr";
			ib.style.fontFamily = "Verdana";
			i.appendChild(ib);
		}
	}

	/***********************************************************/
	//  Presentation helpers
	/***********************************************************/
	function PrintLink(
		parent,
		radius,
		stindex,
		arrow,
		str,
		nclass,
		ncptr,
		chap,
		ctx,
	)
	{
		if (arrow == null)
		{
			arrow = "broken arrow";
		}

		// BEGIN column_radiusof3 i.e. 2of3 and 3of3 depending on radius variable
		let box = document.createElement("div");
		box.id = "radius-" + radius;

		// any arrow comes first
		let prefix = " . .  " + "╠═▹  ";

		for (let indent = 0; indent < radius; indent++)
		{
			prefix = " . . . . . " + prefix;
		}

		if (radius == 2)
		{
			prefix = " . . . . . . .  ║ . . .  " + prefix;
		}

		let hier_pre = document.createElement("span");
		hier_pre.textContent = prefix;
		box.appendChild(hier_pre);

		let arrow_link = document.createElement("a");
		arrow_link.textContent = " ( " + arrow + " )  ";
		arrow_link.id = "arrow-" + stindex;
		arrow_link.title = STINDICES[stindex];
		arrow_link.class = "tooltip";
		arrow_link.style.fontFamily = "Verdana";
		box.appendChild(arrow_link);

		if (str.includes("\n"))
		{
			// preformatted item -> <pre>
			// pre formatted text
			let text_link = document.createElement("a");
			text_link.onclick = function ()
			{
				sendlinkData(nclass, ncptr);
			};

			let pretxt = document.createElement("pre");
			pretxt.textContent = str;

			text_link.appendChild(pretxt);
			text_link.className = "text";
			box.appendChild(text_link);
		} else
		{
			// now plain text
			let text_link = document.createElement("a");
			let spantext = document.createElement("span"); // replaces <pre>

			if (IsURL(str, arrow))
			{
				text_link.href = str;
				text_link.target = "_blank";
				text_link.rel = "noopener";
			} else if (IsImage(str, arrow))
			{
				let img = document.createElement("img");
				img.src = str;
				box.appendChild(img);
			} else
			{
				// THIS IS WHERE WE WANT TO ADD TAB/ORBIT POPUP ON MOUSEOVER ??
				text_link.onclick = function ()
				{
					sendlinkData(nclass, ncptr);
				};
			}

			spantext.textContent = str;
			spantext.style.fontFamily = "Times"; // distinguish satellites
			if (str.length < 20)
			{
				text_link.style.fontSize = "200%";
			}

			text_link.appendChild(spantext);
			box.appendChild(text_link);

			ProgressCheckBox(box, nclass, ncptr, chap, ctx);
		}

		if (ctx.length > 0)
		{
			let cntx = document.createElement("i");
			cntx.id = "cntxt";
			cntx.textContent = " context hints: " + ctx;
			box.appendChild(cntx);
		}

		parent.appendChild(box);
		return parent;
	}

	/***********************************************************/
	function PrintPaths(parent, array)
	{
		if (array.length < 1)
		{
			return parent;
		}

		let lastx = 0;
		let lasty = 0;
		let lastz = 0;
		let lastarrow = 0;
		let newpath;

		for (let path = 0; path < array.length; path++)
		{
			if (array[path] == null)
			{
				continue;
			}

			// The WebPath protocol alternates node-arrow...
			for (let i = 0; i < array[path].length; i++)
			{
				if (i == 0)
				{
					newpath = document.createElement("div");
					newpath.className = "card-view";
					parent.appendChild(newpath);
				}

				if (i % 2 == 0)
				{
					// node
					let str = array[path][i].Name;
					let ncptr = array[path][i].NPtr.CPtr;
					let nclass = array[path][i].NPtr.NClass;

					let xyz = array[path][i].XYZ;

					thisx = xyz.X;
					thisy = xyz.Y;
					thisz = xyz.Z;

					DrawPath(lastarrow, thisx, thisy, thisz, lastx, lasty, lastz);

					if (i < array[path].length - 1)
					{
						lastx = thisx;
						lasty = thisy;
						lastz = thisz;
					} else
					{
						lastx = 0;
						lasty = 0;
						lastz = 0;
					}

					Event(thisx, thisy, thisz);
					Label(thisx, thisy, thisz, str.slice(0, 25), 12, "black");

					if (array[path][i].NPtr != null)
					{
						ncptr = array[path][i].NPtr.CPtr;
						nclass = array[path][i].NPtr.Class;
					}

					if (str.includes("\n"))
					{
						let text_link = document.createElement("a");
						text_link.onclick = function ()
						{
							sendlinkData(nclass, ncptr);
						};

						let pre = document.createElement("pre");
						pre.textContent = str;
						text_link.appendChild(pre);

						newpath.appendChild(text_link);
					} else
					{
						let text_link = document.createElement("a");
						text_link.onclick = function ()
						{
							sendlinkData(nclass, ncptr);
						};

						let text = document.createElement("span");
						text.textContent = str;

						if (str.length < 20)
						{
							text.style.fontSize = "150%";
						}

						text_link.appendChild(text);
						newpath.appendChild(text_link);
					}
				} // arrow
				else
				{
					const then = 2; // reserved vectors
					const prev = 3;

					let arrow = array[path][i].Name;
					let arrptr = array[path][i].Arr;
					let stindex = array[path][i].STindex;

					lastarrow = stindex - ST_ZERO;

					if (arrptr == then || arrptr == prev)
					{
						// represent a privileged sequence for proper time
						let newline = document.createElement("p");
						newpath.appendChild(newline);
					}

					let arrow_link = document.createElement("a");
					arrow_link.textContent = `( ${arrow} )  `;
					arrow_link.id = `arrow-` + stindex;
					arrow_link.class = "tooltip";
					arrow_link.title = STINDICES[stindex];
					arrow_link.style.fontFamily = "Verdana";
					newpath.appendChild(arrow_link);
				}
			}
		}

		return parent;
	}

	/***********************************************************/
	function DrawPath(lastarrow, thisx, thisy, thisz, lastx, lasty, lastz)
	{
		if (lastx != 0 || lasty != 0 || lastz != 0)
		{
			switch (lastarrow)
			{
			case -3:
				Expresses(thisx, thisy, thisz, lastx, lasty, lastz);
				break;
			case -2:
				Contains(thisx, thisy, thisz, lastx, lasty, lastz);
				break;
			case -1:
				LeadsTo(thisx, thisy, thisz, lastx, lasty, lastz);
				break;
			case 0:
				LeadsTo(thisx, thisy, thisz, lastx, lasty, lastz);
				break;
			case 1:
				LeadsTo(lastx, lasty, lastz, thisx, thisy, thisz);
				break;
			case 2:
				Contains(lastx, lasty, lastz, thisx, thisy, thisz);
				break;
			case 3:
				Expresses(lastx, lasty, lastz, thisx, thisy, thisz);
				break;
			default:
				console.log("Bad value", lastarrow);
			}
		}
	}

	/***********************************************************/
	function PrintNotes(parent, array)
	{
		RerenderMath();
		if (array == null || array.length < 1)
		{
			return parent;
		}

		// The root panel is called "parent".
		// Each new Item will start a "child" card
		// Inside child cards, there will be sublines
		let lastx = 0;
		let lasty = 0;
		let lastz = 0;
		let sttype = 3;
		let lastline = 0;
		let last_line_start = "";

		let chtxt;
		let child;
		let subline;
		let lastchtxt;

		// Line by line
		for (let line = 0; line < array.length; line++)
		{
			if (array[line] == null)
			{
				continue;
			}

			// The WebPath protocol alternates node-arrow...
			// i = each item on a line gets added to child card,
			// but there can also be multiple lines inside a card, if referring
			// to the same node/item
			for (let i = 0; i < array[line].length; i++)
			{
				// even items are text, odd items are arrows
				if (i % 2 == 0)
				{
					let str = array[line][i].Name;
					let ncptr = array[line][i].NPtr.CPtr;
					let nclass = array[line][i].NPtr.NClass;

					let xyz = array[line][i].XYZ;

					let thisx = xyz.X;
					let thisy = xyz.Y;
					let thisz = xyz.Z;

					// Add graphic rendering
					if (i == 0 && line > 0)
					{
						// Connect lines in order
						let prx = array[lastline][0].XYZ.X;
						let pry = array[lastline][0].XYZ.Y;
						let prz = array[lastline][0].XYZ.Z;
						DrawPath(sttype, prx, pry, prz, thisx, thisy, thisz);
					}

					if (i > 0)
					{
						DrawPath(sttype, lastx, lasty, lastz, thisx, thisy, thisz);
					}

					lastx = thisx;
					lasty = thisy;
					lastz = thisz;

					Event(thisx, thisy, thisz);
					Label(thisx, thisy, thisz, str.slice(0, 25), 12, "black");

					if (array[line][i].NPtr != null)
					{
						ncptr = array[line][i].NPtr.CPtr;
						nclass = array[line][i].NPtr.Class;
					}

					// if the last line starts with the same item, don't repeat it, use ditto
					if (i == 0 && str == last_line_start)
					{
						// repeated item, so ditto
						subline = document.createElement("div");
						child.append(subline);
						let text = document.createElement("span");
						text.id = "ditto";
						text.textContent = ' . . . .  .   " . . . . .   ';
						text.style.fontFamily = "Times New Roman";
						subline.appendChild(text);
						continue;
					} else if (i == 0)
					{
						// New line has possble new chap/context
						lastline = line;

						let line_no = document.createElement("span");

						line_no.textContent = "At line " + array[line][i].Line;
						parent.appendChild(line_no);

						chtxt = array[line][i].Chp + ":" + array[line][i].Ctx;

						if (chtxt.length > 4 && chtxt != lastchtxt)
						{
							let sec = document.createElement("i");
							sec.textContent = '  From: "' + chtxt + '"';
							parent.appendChild(sec);
						}

						lastchtxt = chtxt;

						child = document.createElement("div");
						child.className = "card-view";
						parent.appendChild(child);
						subline = document.createElement("div");
						child.append(subline);
					}

					if (i == 0)
					{
						// update current
						last_line_start = str;
					}

					// Subsequent items on same line inside a card
					// if pre-formatted
					if (str.includes("\n"))
					{
						let text_link = document.createElement("a");
						text_link.onclick = function ()
						{
							sendlinkData(nclass, ncptr);
						};
						let pre = document.createElement("pre");
						pre.textContent = str;
						text_link.appendChild(pre);
						subline.appendChild(text_link);
					} else
					{
						let text_link = document.createElement("a");
						text_link.onclick = function ()
						{
							sendlinkData(nclass, ncptr);
						};

						let text = document.createElement("span");
						text.textContent = str;
						text_link.appendChild(text);
						subline.appendChild(text_link);
					}

					ProgressCheckBox(
						subline,
						nclass,
						ncptr,
						array[line][i].Chp,
						array[line][i].Ctx,
					);
				} // arrow
				else
				{
					const then = 2; // reserved vectors
					const prev = 3;

					let arrow = array[line][i].Name;
					let arrptr = array[line][i].Arr;
					let stindex = array[line][i].STindex;

					sttype = stindex - ST_ZERO;

					if (arrptr == then || arrptr == prev)
					{
						// represent a privileged sequence for proper time
						let newline = document.createElement("p");
						subline.appendChild(newline);
					}

					let arrow_link = document.createElement("a");
					arrow_link.textContent = `( ${arrow} )  `;
					arrow_link.id = `arrow-` + stindex;
					arrow_link.class = "tooltip";
					arrow_link.title = STINDICES[stindex];
					arrow_link.style.fontFamily = "Verdana";

					subline.appendChild(arrow_link);
				} // if not i % 2 = arrow
			} // for each item i
		} // for each line

		return parent;
	}

	/***********************************************************/
	function ShowNodeEvent(
		panel,
		event,
		counter,
		direction,
		skiparrow,
		anchortag,
	)
	{
		// create a div panel for the orbit root node
		let child = document.createElement("div");
		child.className = "card-view";
		child.id = "orbit_column_1of3"; // col 1 spans all 3 columns, full width for full node text
		panel.appendChild(child);

		if (event == null)
		{
			return;
		}

		let text = counter + ". " + event.Text;
		let nptrtxt = "(" + event.NPtr.Class + "," + event.NPtr.CPtr + ")";

		if (counter == 0)
		{
			text = "--> " + event.Text;
		}

		// ** BEGIN 1: Here we print the full text for column_1of3 either as pre or p
		if (text.includes("\n"))
		{
			let from_link = document.createElement("a");
			from_link.onclick = function ()
			{
				sendlinkData(event.NPtr.Class, event.NPtr.CPtr);
			};

			let from_text = document.createElement("pre");
			from_text.nameClass = "text";
			from_text.textContent = text;
			from_link.nameClass = "text";
			from_link.appendChild(from_text);

			child.appendChild(from_link);
		} else
		{
			let from_link = document.createElement("span");
			from_link.onclick = function ()
			{
				sendlinkData(event.NPtr.Class, event.NPtr.CPtr);
			};
			let from_text = document.createElement(anchortag);
			from_link.nameClass = "text";
			from_link.appendChild(from_text);

			child.appendChild(from_link);

			if (!IsMath(event.Text))
			{
				from_text.textContent = event.Text.slice(0, 70) + "...";
				let small_tot_text = document.createElement("div");
				small_tot_text.textContent = text;
				small_tot_text.id = "orbital-full-text";
				child.appendChild(small_tot_text);
			} else
			{
				from_text.textContent = text; // event.Text;
			}
		}

		// If this is the root node, we need to add some Nptr, context info in column_1of3
		if (counter == 0)
		{
			let setting = document.createElement("span");
			setting.id = "nptr-chapter-context-helpline";

			let text1 = document.createElement("i");
			text1.textContent = "with NPtr " + nptrtxt + ", in chapter ";
			setting.appendChild(text1);

			let chplink = document.createElement("a");
			chplink.textContent = '"' + event.Chap + '"';
			chplink.onclick = function ()
			{
				sendLinkSearch('any \\chapter "' + event.Chap + '"');
			};
			setting.appendChild(chplink);

			let text2 = document.createElement("i");
			text2.textContent = ", context ";
			setting.appendChild(text2);

			let ctxlink = document.createElement("a");
			ctxlink.textContent = '"' + event.Context + '"';

			ctxlink.onclick = function ()
			{
				sendLinkSearch('any \\context "' + CtxSplice(event.Context) + '"');
			};
			setting.appendChild(ctxlink);

			child.appendChild(setting);
		}

		ProgressCheckBox(
			child,
			event.NPtr.Class,
			event.NPtr.CPtr,
			event.Chap,
			event.Context,
		);

		// See what pathways we are part of and add notes
		CheckSingleCone(
			panel,
			child,
			"[LT]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			1,
			event.Orbits[Im1],
			event.Orbits[Il1],
		);
		CheckSingleCone(
			panel,
			child,
			"[CN]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			2,
			event.Orbits[Im2],
			event.Orbits[Ic2],
		);
		CheckSingleCone(
			panel,
			child,
			"[EP]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			3,
			event.Orbits[Im3],
			event.Orbits[Ie3],
		);
		CheckSingleCone(
			panel,
			child,
			"[NR]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			0,
			event.Orbits[In0],
			event.Orbits[In0],
		);

		// END : column_1of3
		// Next we add the satellite nodes in column_2of3 and column_3of3
		// Each vector from Im3...Ie3, one by one
		AddLinkOrbits(panel, child, event, Im3, skiparrow);
		AddLinkOrbits(panel, child, event, Ie3, skiparrow);

		// Then any equivalents
		AddLinkOrbits(panel, child, event, In0, skiparrow);

		// Then containment
		AddLinkOrbits(panel, child, event, Im2, skiparrow);
		AddLinkOrbits(panel, child, event, Ic2, skiparrow);

		// Lastly, from-->to
		AddLinkOrbits(panel, child, event, Im1, skiparrow);
		AddLinkOrbits(panel, child, event, Il1, skiparrow);
	}

	/***********************************************************/
	function ShowSequenceItem(
		panel,
		event,
		counter,
		direction,
		skiparrow,
		anchortag,
	)
	{
		// create a div panel for the orbit root node
		let child = document.createElement("div");
		child.className = "card-view";
		panel.appendChild(child);

		if (event == null)
		{
			return;
		}

		let text = counter + ". " + event.Text;
		let nptrtxt = "(" + event.NPtr.Class + "," + event.NPtr.CPtr + ")";

		// ** BEGIN 1: Here we print the full text for column_1of3 either as pre or p
		if (text.includes("\n"))
		{
			let from_link = document.createElement("a");
			from_link.onclick = function ()
			{
				sendlinkData(event.NPtr.Class, event.NPtr.CPtr);
			};

			let from_text = document.createElement("pre");
			from_text.nameClass = "text";
			from_text.textContent = text;
			from_link.nameClass = "text";
			from_link.appendChild(from_text);

			child.appendChild(from_link);
		} else
		{
			let from_link = document.createElement("span");
			from_link.onclick = function ()
			{
				sendlinkData(event.NPtr.Class, event.NPtr.CPtr);
			};
			let from_text = document.createElement(anchortag);
			from_link.nameClass = "text";
			from_link.appendChild(from_text);

			child.appendChild(from_link);

			from_text.textContent = text + "    ";
		}

		ProgressCheckBox(
			child,
			event.NPtr.Class,
			event.NPtr.CPtr,
			event.Chap,
			event.Context,
		);

		// See what pathways we are part of and add notes
		CheckSingleCone(
			panel,
			child,
			"[LT]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			1,
			event.Orbits[Im1],
			event.Orbits[Il1],
		);
		CheckSingleCone(
			panel,
			child,
			"[CN]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			2,
			event.Orbits[Im2],
			event.Orbits[Ic2],
		);
		CheckSingleCone(
			panel,
			child,
			"[EP]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			3,
			event.Orbits[Im3],
			event.Orbits[Ie3],
		);
		CheckSingleCone(
			panel,
			child,
			"[NR]",
			event.NPtr.Class,
			event.NPtr.CPtr,
			0,
			event.Orbits[In0],
			event.Orbits[In0],
		);

		let text2 = document.createElement("i");
		text2.textContent = ", . . . context ";
		child.appendChild(text2);

		let ctxlink = document.createElement("a");
		ctxlink.textContent = '"' + event.Context + '"';
		ctxlink.onclick = function ()
		{
			sendLinkSearch('any \\context "' + CtxSplice(event.Context) + '"');
		};
		child.appendChild(ctxlink);
	}

	/***********************************************************/
	function ProgressCheckBox(container, nclass, ncptr, chap, context)
	{
		// Create a label for the checkbox
		let label = document.createElement("label");
		label.className = "checkbox";
		label.textContent = "";

		let checkbox = document.createElement("input");
		checkbox.type = "checkbox";
		checkbox.id = "progress-checkbox";
		checkbox.value = "(" + nclass + "," + ncptr + ")";

		checkbox.addEventListener("change", function ()
		{
			let chapcontext = chap + ":" + context;
			if (checkbox.checked)
			{
				let ack = sendlastseen("\\lastnptr", nclass, ncptr, chapcontext);
			}
		});

		let spanner = document.createElement("span");
		spanner.className = "checkmark";

		label.appendChild(checkbox);
		label.appendChild(spanner);
		container.appendChild(label);
	}

	/***********************************************************/
	// Busy wheel of time...
	/***********************************************************/
	/***********************************************************/
	// Render line link list by STtype
	/***********************************************************/
	function AddLinkOrbits(panel, child, event, sttype, skiparrow)
	{
		if (event.Orbits[sttype] != null)
		{
			for (let ngh of event.Orbits[sttype])
			{
				if (skiparrow != ngh.Arrow)
				{
					child = PrintLink(
						child,
						ngh.Radius,
						ngh.STindex,
						ngh.Arrow,
						ngh.Text,
						ngh.Dst.Class,
						ngh.Dst.CPtr,
						event.Chap,
						ngh.Ctx,
					);
					panel.appendChild(child);

					if (sttype == Im3 && IsImage(event.Text, ngh.Arrow))
					{
						let img = document.createElement("img");
						img.src = event.Text;
						panel.appendChild(img);
					}
				}
			}
		}
	}

	function CheckCones(panel, docpart, name, bwd, fwd)
	{
		if (bwd[0] != null && fwd[0] != null)
		{
			// we are in the middle of a cone
			let nptrs = "";

			for (let b of bwd)
			{
				let np = "(" + b.Dst.Class + "," + b.Dst.CPtr + ") ";
				nptrs = nptrs + np;
			}

			for (let f of fwd)
			{
				let np = "(" + f.Dst.Class + "," + f.Dst.CPtr + ") ";
				nptrs = nptrs + np;
			}

			let link = document.createElement("a");
			link.textContent = name;
			link.onclick = function ()
			{
				sendLinkSearch("\\from " + nptrs + " limit 30");
			};
			docpart.id = "cone_button";
			docpart.appendChild(link);
		}

		panel.appendChild(docpart);
	}

	/***********************************************************/
	function CheckSingleCone(
		panel,
		docpart,
		name,
		nclass,
		nptr,
		arrow,
		bwd,
		fwd,
	)
	{
		if (bwd == null || fwd == null)
		{
			return;
		}

		if (bwd[0] != null && fwd[0] != null)
		{
			// we are in the middle of a cone
			let link = document.createElement("a");
			link.textContent = name;
			link.onclick = function ()
			{
				sendLinkSearch(
					"\\from " +
					"(" +
					nclass +
					"," +
					nptr +
					") \\arrow " +
					arrow +
					" \\limit 30",
				);
			};

			docpart.id = "cone_button";
			docpart.appendChild(link);
		}
	}

	/***********************************************************/
	function IsImage(str, arrow)
	{
		if (arrow == "has image" || arrow == "is an image for")
		{
			if (str.slice(0, 6) == "http:/" || str.slice(0, 7) == "https:/")
			{
				return true;
			}
		}

		return false;
	}

	/***********************************************************/
	function IsMath(str)
	{
		if (str.includes("(") && str.includes(")"))
		{
			return true;
		}

		return false;
	}

	/***********************************************************/
	function IsURL(str, arrow)
	{
		if (arrow == "has URL")
		{
			if (str.slice(0, 6) == "http:/" || str.slice(0, 7) == "https:/")
			{
				return true;
			}
		}

		return false;
	}

	/***********************************************************/
	// handlers
	/***********************************************************
    
	 
	 /***********************************************************/
	function SearchHandler()
	{
		const form = document.querySelector("#search");
		if (!form) return;

		async function sendsearchData()
		{
			const formData = new FormData(form);
			const searchQuery = formData.get("name");

			saveSearchToHistory(searchQuery);
			const state = { searchQuery: searchQuery };
			const title = "Results for: " + searchQuery;
			const url =
				window.location.pathname + "?search=" + encodeURIComponent(searchQuery);
			pushStateSafe(state, title, url);
			startHipnotize();

			fetch("/searchN4L", { method: POST_METHOD, body: formData })
				.then((response) =>
				{
					stopHipnotize();
					if (!response.ok)
					{
						DisplayError("SearchHandler() - network returns error");
						throw new Error("network returns error");
					}
					return response.json();
				})
				.then((resp) =>
				{
					DoHeader(resp);

					switch (resp.Response)
					{
					case "Orbits":
						DoOrbitPanel(resp);
						break;
					case "ConePaths":
						DoEntireConePanel(resp);
						break;
					case "PathSolve":
						DoEntireConePanel(resp);
						break;
					case "Sequence":
						DoSeqPanel(resp);
						break;
					case "PageMap":
						DoPageMapPanel(resp);
						break;
					case "TOC":
						DoTOCPanel(resp);
						break;
					case "Arrows":
						DoArrowsPanel(resp);
						break;
					case "STAT":
						DoStatsPanel(resp);
						break;
					}

					const indicator = document.getElementById("scroll-indicator");
					if (indicator)
					{
						indicator.classList.add("visible");
					}
				})
				.catch((error) =>
				{
					console.log("error ", error);
					DisplayError("No results (perhaps no connection)");
				});
		}

		// Take over form submission
		const button = document.getElementById("gosubmit");
		if (button)
		{
			button.addEventListener("click", (event) =>
			{
				event.preventDefault();
				sendsearchData();
				topFunction(); // Assuming topFunction() scrolls to top
			});
		}
	}

	/***********************************************************/
	async function sendlinkData(nclass, ncptr)
	{
		let formData = new FormData();
		formData.set("nclass", nclass);
		formData.set("ncptr", ncptr);

		let searchfield = document.getElementById("name");
		searchfield.value = "(" + nclass + "," + ncptr + ")";
		topFunction();
		startHipnotize();

		fetch("/searchN4L", { method: POST_METHOD, body: formData })
			.then((response) =>
			{
				stopHipnotize();

				if (!response.ok)
				{
					DisplayError("sendlinkData - Response Error");
					throw new Error("network returns error");
				}

				return response.json();
			})
			.then((resp) =>
			{
				DoHeader(resp);
				DoOrbitPanel(resp);
			})

			.catch((error) =>
			{
				console.log("error ", error);
				DisplayError("sendlinkData", err);
			});
	}

	/***********************************************************/
	async function sendlastseen(name, nclass, ncptr, chapcontext)
	{
		let formData = new FormData();
		formData.set("nclass", nclass);
		formData.set("ncptr", ncptr);
		formData.set("name", name);
		formData.set("chapcontext", chapcontext);

		fetch("/searchN4L", { method: POST_METHOD, body: formData })
			.then((response) =>
			{
				if (!response.ok)
				{
					throw new Error("network returns error");
				}

				return response.json();
			})

			.then((resp) =>
			{
				console.log("LASTSAW ACK", resp.Content);
				return;
			})

			.catch((error) =>
			{
				// Handle error
				console.log("error ", error);
			});
	}

	/***********************************************************/
	async function sendLinkSearch(search)
	{
		let formData = new FormData();
		formData.set("name", search);

		let searchfield = document.getElementById("name");
		searchfield.value = search;

		startHipnotize();

		fetch("/searchN4L", { method: POST_METHOD, body: formData })
			.then((response) =>
			{
				if (!response.ok)
				{
					throw new Error("network returns error");
				}

				return response.json();
			})

			.then((resp) =>
			{
				const state = { searchQuery: search };
				const title = "Results for: " + search;
				// We explicitly include the current path to be browser-compliant.
				const url =
					window.location.pathname + "?search=" + encodeURIComponent(search);
				pushStateSafe(state, title, url);

				stopHipnotize();

				DoHeader(resp);

				switch (resp.Response)
				{
				case "Orbits":
					DoOrbitPanel(resp);
					break;
				case "ConePaths":
					DoEntireConePanel(resp);
					break;
				case "PathSolve":
					DoEntireConePanel(resp);
					break;
				case "Sequence":
					DoSeqPanel(resp);
					break;
				case "PageMap":
					DoPageMapPanel(resp);
					break;
				case "TOC":
					DoTOCPanel(resp);
					break;
				case "Arrows":
					DoArrowsPanel(resp);
					break;
				}
			})

			.catch((error) =>
			{
				console.log("error ", error);
			});
	}

	/***********************************************************/
	function CtxSplice(s)
	{
		let ret = s.replaceAll(" . ", ".");
		return ret;
	}

	/***********************************************************/
	// Graphics Drawing
	/***********************************************************/
	function HeatColour(freq, pdelta, sat)
	{
		// pdelta is measured in seconds --> HSL
		const hottest = 100;
		const coldest = 3600 * 24 * 7 - hottest;

		let hue = ((pdelta - hottest) / coldest) * 230;

		let saturation = sat;

		let lightness = 40 + freq * 2;

		if (hue < 0)
		{
			hue = 0;
		}

		if (hue > 230)
		{
			hue = 230;
		}

		if (lightness > 100)
		{
			lightness = 100;
		}

		let hsl = `hsl(${hue}, ${saturation}%, ${lightness}%)`;

		return hsl;
	}

	/***********************************************************/

	function PlotGraphics(event, lastevent)
	{
		let tx = event.XYZ.X;
		let ty = event.XYZ.Y;
		let tz = event.XYZ.Z;

		Event(tx, ty, tz);
		Label(tx, ty, tz, event.Text.slice(0, 25), 12, "black");

		if (lastevent != event)
		{
			let lx = lastevent.XYZ.X;
			let ly = lastevent.XYZ.Y;
			let lz = lastevent.XYZ.Z;

			LeadsTo(lx, ly, lz, tx, ty, tz);
		}

		// Now look at orbits

		if (event.Orbits[Il1] != null)
		{
			for (let ngh of event.Orbits[Il1])
			{
				Event(ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
				LeadsTo(
					ngh.OOO.X,
					ngh.OOO.Y,
					ngh.OOO.Z,
					ngh.XYZ.X,
					ngh.XYZ.Y,
					ngh.XYZ.Z,
				);
			}
		}

		if (event.Orbits[Im1] != null)
		{
			for (let ngh of event.Orbits[Im1])
			{
				Event(ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
				LeadsTo(
					ngh.XYZ.X,
					ngh.XYZ.Y,
					ngh.XYZ.Z,
					ngh.OOO.X,
					ngh.OOO.Y,
					ngh.OOO.Z,
				);
			}
		}

		if (event.Orbits[Ic2] != null)
		{
			for (let ngh of event.Orbits[Ic2])
			{
				Thing(ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
				Contains(
					ngh.OOO.X,
					ngh.OOO.Y,
					ngh.OOO.Z,
					ngh.XYZ.X,
					ngh.XYZ.Y,
					ngh.XYZ.Z,
				);
			}
		}

		if (event.Orbits[Im2] != null)
		{
			for (let ngh of event.Orbits[Im2])
			{
				Thing(ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
				Contains(
					ngh.XYZ.X,
					ngh.XYZ.Y,
					ngh.XYZ.Z,
					ngh.OOO.X,
					ngh.OOO.Y,
					ngh.OOO.Z,
				);
			}
		}

		if (event.Orbits[Ie3] != null)
		{
			for (let ngh of event.Orbits[Ie3])
			{
				Concept(ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
				Expresses(
					ngh.OOO.X,
					ngh.OOO.Y,
					ngh.OOO.Z,
					ngh.XYZ.X,
					ngh.XYZ.Y,
					ngh.XYZ.Z,
				);
			}
		}

		if (event.Orbits[Im3] != null)
		{
			for (let ngh of event.Orbits[Im3])
			{
				Concept(ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
				Expresses(
					ngh.XYZ.X,
					ngh.XYZ.Y,
					ngh.XYZ.Z,
					ngh.OOO.X,
					ngh.OOO.Y,
					ngh.OOO.Z,
				);
			}
		}

		if (event.Orbits[In0] != null)
		{
			for (let ngh of event.Orbits[In0])
			{
				Event(ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
				Near(ngh.OOO.X, ngh.OOO.Y, ngh.OOO.Z, ngh.XYZ.X, ngh.XYZ.Y, ngh.XYZ.Z);
			}
		}
	}

	/***********************************************************/
	function ArrowPair(angle, type, fwd, bwd)
	{
		x = 0.5 * Math.cos(angle);
		y = 0.5 * Math.sin(angle);

		switch (type)
		{
		case 0:
			Near(0, 0, 0, x, y, 0);
			Near(0, 0, 0, -x, -y, 0);
			break;
		case 1:
		case -1:
			LeadsTo(0, 0, 0, x, y, 0);
			LeadsTo(0, 0, 0, -x, -y, 0);
			break;
		case 2:
		case -2:
			Contains(0, 0, 0, x, y, 0);
			Contains(0, 0, 0, -x, -y, 0);
			break;
		case 3:
		case -3:
			Expresses(0, 0, 0, x, y, 0);
			Expresses(0, 0, 0, -x, -y, 0);
			break;
		}

		const size = "12";
		const colour = "black";

		Label(x, y, 0, fwd, size, colour);
		Label(-x, -y, 0, bwd, size, colour);
	}

	/***********************************************************/
	/* GRAPHICS PANE                                           */
	/***********************************************************/
	function DrawWelcomeImage()
	{
		DrawGrid(0, 0, 1);
		return;
		let orbit = 0.5;
		let x0 = 0;
		let y0 = 0;

		for (let z = 1; z > -1.0; z -= orbit)
		{
			LeadsTo(x0, y0, z, 0, 0, z + orbit);
			Event(x0, y0, z, 10);

			Label(x0, y0, z, "SST event " + z, 16, "darkblue");

			for (let a = 0; a < 2 * Math.PI; a += Math.PI / 6)
			{
				let x = orbit * Math.cos(a);
				let y = orbit * Math.sin(a);
				Concept(x, y, z, 6);
				Expresses(0, 0, z, x, y, z);
			}
		}
	}

	/***********************************************************/
	function DrawGrid(x, z, length)
	{
		CTX.save();

		for (let xi = -length; xi <= length; xi += 0.1)
		{
			SST_Line(xi, 0, -length, xi, 0, length, "lightgrey", 1 * mob);
		}

		for (let zi = -length; zi <= length; zi += 0.1)
		{
			SST_Line(-length, 0, zi, length, 0, zi, "lightgrey", 1 * mob);
		}

		SST_Line(-length / 2, 0, 0, 0, 0, 0, "lightgrey", 1 * mob);
		SST_Line(0, 0, -length / 2, 0, 0, 0, "lightgrey", 1 * mob);
		SST_Line(0, -length / 2, 0, 0, length, 0, "lightgrey", 1 * mob);

		CTX.restore();
	}

	// *************************************************
	function CreateCanvas()
	{
		let oldcanvas = document.getElementById("myCanvas");

		if (oldcanvas != null)
		{
			oldcanvas.remove();
		}
		let parent = document.getElementById("canvas");
		let canvas = document.createElement("canvas");
		canvas.id = "myCanvas";
		canvas.width = window.innerWidth;
		canvas.height = window.innerHeight;
		CTX = canvas.getContext("2d");
		CTX.beginPath();

		parent.appendChild(canvas);
		return canvas;
	}

	async function resizeCanvas()
	{
		let canvas = document.getElementById("myCanvas");
		let scale = Math.min(
			window.innerWidth / canvas.width,
			window.innerHeight / canvas.height,
		);

		canvas.style.width = `${Math.round(scale * canvas.width)}px`;
		canvas.style.height = `${Math.round(scale * canvas.height)}px`;
	}

	// *************************************************
	function Label(x, y, z, text, size, colour)
	{
		CTX.save();
		let font = "bold " + size * mob + "px sans-serif";

		let xr = Tx(x, y, z) + 30;
		let yr = Ty(x, y, z);

		CTX.beginPath();

		let w = CTX.measureText(text).width;
		let h = parseInt(font, size);
		CTX.fillStyle = "transparent";
		CTX.font = font;
		CTX.fillStyle = colour;
		CTX.fillText(text, xr, yr);
		CTX.restore();
	}

	// *************************************************
	function Horizon(x, y, z)
	{
		return Math.sqrt(
			(x - OBS_X) * (x - OBS_X) +
			(y - OBS_Y) * (y - OBS_Y) +
			(z - OBS_Z) * (z - OBS_Z),
		);
	}

	// *************************************************
	function Alpha(x, y, z)
	{
		let alpha =
			1.5 /
			Math.sqrt(
				(x - OBS_X) * (x - OBS_X) +
				(y - OBS_Y) * (y - OBS_Y) +
				(z - OBS_Z) * (z - OBS_Z),
			);

		if (alpha > 1)
		{
			return 1;
		}

		if (alpha < 0.1)
		{
			return 0.1;
		}
	}
	// *************************************************
	function Tx(x, y, z)
	{
		let scale = (SCALE * WIDTH) / (1 + 1.2 * Horizon(x, y, z));

		let xt = ORGX + scale * (x * Math.cos(THETA) + z * Math.cos(PHI));
		return xt;
	}

	// *************************************************
	function Ty(x, y, z)
	{
		let scale = (SCALE * WIDTH) / (1 + 1.5 * Horizon(x, y, z));
		let yt =
			HEIGHT - ORGY - scale * (y + z * Math.sin(PHI) - x * Math.sin(THETA));
		return yt;
	}

	// *************************************************
	function LeadsTo(x0, y0, z0, xp, yp, zp)
	{
		//Arrow(x0,y0,z0,xp,yp,zp,"rgba(0,250,0,1)",3);
		Arrow(x0, y0, z0, xp, yp, zp, "darkred", 3 * mob);
	}

	// *************************************************
	function Contains(x0, y0, z0, xp, yp, zp)
	{
		//Arrow(x0,y0,z0,xp,yp,zp,"rgba(60,60,60,1)",2);
		Arrow(x0, y0, z0, xp, yp, zp, "lightblue", 2 * mob);
	}

	// *************************************************
	function Expresses(x0, y0, z0, xp, yp, zp)
	{
		//Arrow(x0,y0,z0,xp,yp,zp,"rgba(106,236,255,1)",2);
		Arrow(x0, y0, z0, xp, yp, zp, "orange", 2 * mob);
	}

	// *************************************************
	function Near(x0, y0, z0, xp, yp, zp)
	{
		//Arrow(x0,y0,z0,xp,yp,zp,"rgba(20,20,20,1)",1);
		Arrow(x0, y0, z0, xp, yp, zp, "darkgrey", 1 * mob);
	}

	// *************************************************
	function Event(x, y, z)
	{
		Node(x, y, z, 6 * mob, "darkred", "red");
	}

	// *************************************************
	function Thing(x, y, z)
	{
		Node(x, y, z, 4 * mob, "darkgreen", "lightgreen");
	}

	// *************************************************
	function Concept(x, y, z)
	{
		Node(x, y, z, 4 * mob, "darkblue", "lightblue");
	}

	// *************************************************
	function Node(x, y, z, r, col1, col2)
	{
		CTX.save();
		CTX.beginPath();
		let x0 = Tx(x, y, z);
		let y0 = Ty(x, y, z);
		r = (r * 1.6) / Horizon(x, y, z);

		let grad = CTX.createLinearGradient(x0, y0, x0 + r, y0 + r);

		grad.addColorStop(0, col2);
		grad.addColorStop(1, col1);

		//CTX.globalAlpha = 1-Alpha(x,y,x)/3;
		CTX.arc(x0, y0, r, 0, Math.PI * 2);
		CTX.fillStyle = grad;
		CTX.fill();
		CTX.restore();
	}

	// *************************************************
	function Arrow(x0, y0, z0, xp, yp, zp, colour, thickness)
	{
		CTX.save();
		SST_Line(x0, y0, z0, xp, yp, zp, colour, thickness);

		let frx = Tx(x0, y0, z0);
		let fry = Ty(x0, y0, z0);
		let tox = Tx(xp, yp, zp);
		let toy = Ty(xp, yp, zp);
		let scale = 1.1 - zp;
		let angle = Math.atan2(toy - fry, tox - frx);
		let headangle = Math.PI / 12;
		let headlen = 12 * scale;
		let noderadius = 10 * scale;

		CTX.beginPath();
		CTX.strokeStyle = colour;
		CTX.lineWidth = thickness;
		CTX.moveTo(
			tox - noderadius * Math.cos(angle),
			toy - noderadius * Math.sin(angle),
		);
		CTX.lineTo(
			tox - headlen * Math.cos(angle - headangle),
			toy - headlen * Math.sin(angle - headangle),
		);
		CTX.moveTo(
			tox - noderadius * Math.cos(angle),
			toy - noderadius * Math.sin(angle),
		);
		CTX.lineTo(
			tox - headlen * Math.cos(angle + headangle),
			toy - headlen * Math.sin(angle + headangle),
		);
		CTX.stroke();
		CTX.beginPath();
		CTX.restore();
	}

	// *************************************************
	function SST_Line(x0, y0, z0, xp, yp, zp, colour, thickness)
	{
		CTX.save();
		CTX.beginPath();
		let xb = Tx(x0, y0, z0);
		let yb = Ty(x0, y0, z0);
		let xe = Tx(xp, yp, zp);
		let ye = Ty(xp, yp, zp);

		//CTX.globalAlpha = 1;
		CTX.moveTo(xb, yb);
		CTX.lineTo(xe, ye);
		CTX.strokeStyle = colour;
		CTX.lineWidth = thickness;
		CTX.stroke();
		CTX.closePath();
		CTX.beginPath();
		CTX.restore();
	}

	// Waiting Effect functions *****************************************************************
	function startHipnotize()
	{
		const el = document.getElementById("wait");
		if (el)
		{
			el.innerHTML = `<div class="hipnosis">
    <svg version="1.1" id="L1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 100 100" enable-background="new 0 0 100 100" xml:space="preserve">
      <circle fill="none" stroke="#fff" stroke-width="6" stroke-miterlimit="15" stroke-dasharray="14.2472,14.2472" cx="50" cy="50" r="47">
<animateTransform attributeName="transform" attributeType="XML" type="rotate" dur="5s" from="0 50 50" to="360 50 50" repeatCount="indefinite" />
      </circle>
      <circle fill="none" stroke="#fff" stroke-width="1" stroke-miterlimit="10" stroke-dasharray="10,10" cx="50" cy="50" r="39">
        <animateTransform attributeName="transform" attributeType="XML" type="rotate" dur="5s" from="0 50 50" to="-360 50 50" repeatCount="indefinite" />
      </circle>
      <g fill="#fff">
        <rect x="30" y="35" width="5" height="30">
          <animateTransform attributeName="transform" dur="1s" type="translate" values="0 5 ; 0 -5; 0 5" repeatCount="indefinite" begin="0.1" />
        </rect>
        <rect x="40" y="35" width="5" height="30">
          <animateTransform attributeName="transform" dur="1s" type="translate" values="0 5 ; 0 -5; 0 5" repeatCount="indefinite" begin="0.2" />
        </rect>
        <rect x="50" y="35" width="5" height="30">
          <animateTransform attributeName="transform" dur="1s" type="translate" values="0 5 ; 0 -5; 0 5" repeatCount="indefinite" begin="0.3" />
        </rect>
        <rect x="60" y="35" width="5" height="30">
          <animateTransform attributeName="transform" dur="1s" type="translate" values="0 5 ; 0 -5; 0 5" repeatCount="indefinite" begin="0.4" />
        </rect>
        <rect x="70" y="35" width="5" height="30">
          <animateTransform attributeName="transform" dur="1s" type="translate" values="0 5 ; 0 -5; 0 5" repeatCount="indefinite" begin="0.5" />
        </rect>
      </g>
    </svg>
  </div>`;
		}
	}

	function stopHipnotize()
	{
		const el = document.getElementById("wait");
		if (el)
		{
			el.innerHTML = "";
		}
	}

	window.addEventListener(
		"scroll",
		() =>
		{
			const indicator = document.getElementById("scroll-indicator");
			if (indicator)
			{
				indicator.classList.remove("visible");
			}
		},
		{ once: false },
	);

	// Add this function somewhere in your main.js

	async function checkStatus()
	{
		const serverIndicator = document.getElementById("server-status");
		const dbIndicator = document.getElementById("database-status");

		if (!serverIndicator || !dbIndicator) return;

		try
		{
			const response = await fetch("/status");

			if (!response.ok)
			{
				throw new Error("Server responded with an error");
			}

			const status = await response.json();

			// Update Server Status Indicator
			if (status.server_status === "OK")
			{
				serverIndicator.className = "status-indicator status-ok";
				serverIndicator.setAttribute("data-label", "Server OK");
			} else
			{
				serverIndicator.className = "status-indicator status-error";
				serverIndicator.setAttribute("data-label", "Server Error");
			}

			// Update Database Status Indicator
			if (status.database_status === "OK")
			{
				dbIndicator.className = "status-indicator status-ok";
				dbIndicator.setAttribute("data-label", "Database OK");
			} else
			{
				dbIndicator.className = "status-indicator status-error";
				dbIndicator.setAttribute("data-label", "Database Error");
			}
		} catch (error)
		{
			// If the fetch fails, both are in an error state
			console.error("Status check failed:", error);
			serverIndicator.className = "status-indicator status-error";
			serverIndicator.setAttribute("data-label", "Server Unreachable");
			dbIndicator.className = "status-indicator status-error";
			dbIndicator.setAttribute("data-label", "Database Unreachable");
		}
	}

	/**
	 * A safe wrapper for history.pushState that prevents pushing a new state
	 * if the target URL is the same as the current one.
	 * @param {object} state - The state object to save.
	 * @param {string} title - The new document title.
	 * @param {string} url - The new URL to push.
	 */
	function pushStateSafe(state, title, url)
	{
		// We compare the target URL with the browser's current URL.
		if (
			window.location.pathname +
			window.location.search +
			window.location.hash !==
			url
		)
		{
			history.pushState(state, title, url);
		}
	}
	/***********************************************************/
	// Main program starts here
	/***********************************************************/
	function initializeApp()
	{
		// checkStatus();
		// setInterval(checkStatus, 30000);

		loadHistoryIntoDatalist();
		SearchHandler();
		AppRouter();
	}
	window.onresize = resizeCanvas;
	initializeApp();
});
