using System;
using System.Collections.Generic;
using System.Globalization;
using System.IO;
using System.Text.RegularExpressions;
using YamlDotNet.RepresentationModel;
using Tack.Utils;

namespace Tack
{
	public class Page : DataProvider
	{
		// available directly after construction.
		public string Name { get; protected set; }
		public string DiskPath { get; protected set; }
		public Tacker Tacker { get; protected set; }
		public bool IsFloating { get; protected set; }

		bool inited = false;
		// first available after call to Init()
		public Page Parent { get; protected set; }
		public IList<Page> Siblings { get; protected set; }
		public ISet<string> Assets { get; protected set; }
		public IDictionary<string, object> Variables { get; protected set; }
		public string Template { get; protected set; }

		public Page (Tacker tacker, string realpath)
		{
			Tacker = tacker;
			DiskPath = realpath;
			var fn = Path.GetFileName (realpath);
			Name = Regex.Replace (fn, "^[0-9]+\\.", "");
			IsFloating = !Regex.Match (fn, "^[0-9]+\\.").Success;
		}


		public string Permalink {
			get {
				return (Parent == null ? "" : Parent.Permalink) + "/" +
					(Parent == null && "index".Equals (Name) ? "" : Name);
			}
		}

		public IEnumerable<Page> Ancestors {
			get {
				for (var p = Parent; p != null; p = p.Parent) {
					yield return p;
				}
			}
		}

		public void Init ()
		{
			var parent = Path.GetDirectoryName (DiskPath);
			var siblings = new SortedDictionary<string, Page> ();

			foreach (var i in Tacker.Pages) {
				if (i.DiskPath.Equals (parent)) {
					Parent = i;
				}
				if (parent.Equals (Path.GetDirectoryName (i.DiskPath)) &&
				    i != this && !i.IsFloating) {
					siblings.Add (Path.GetFileName (i.DiskPath), i);
				}
			}

			Siblings = new List<Page> (siblings.Values);

			var metadata = new Dictionary<string, object> ();
			var assets = new HashSet<string> ();

			foreach (var i in Directory.GetFiles (DiskPath, "*")) {
				var md = Tacker.ProcessMetadata (i);
				if (md != null) {
					Template = Template ?? Path.GetFileNameWithoutExtension (i);
					metadata.AddAll (md);
					continue;
				}
				var markup = Tacker.ProcessMarkup (i);
				if (markup != null) {
					metadata.Add (Path.GetFileNameWithoutExtension (i), markup);
					continue;
				}

				assets.Add (i.Replace (DiskPath, ""));
			}

			Assets = assets;
			Variables = metadata;
			inited = true;
		}

		public void Generate ()
		{
			if (!inited) {
				Init ();
			}

			Console.WriteLine ("Generating {0}", Name);
			Console.WriteLine (" - ancestors: {0}", String.Join (" << ", Ancestors));
			Console.WriteLine (" - siblings: {0}", String.Join (", ", Siblings));

			Directory.CreateDirectory (Tacker.TargetDir + Permalink);

			using (var writer = File.CreateText(Path.Combine (Tacker.TargetDir + Permalink, "index.html"))) {
				Tacker.FindTemplate (Template).Render (new DictWrapper (this, new RenderContext (this)), writer, Tacker.FindTemplate);
			}
		}

		public override string ToString()
		{
			return Name;
		}

		public object GetData (string key, RenderContext ctx) {
			key = key.Trim ();
			switch (key) {
			case "permalink": return Permalink;
			case "slug": return Name;
			case "name": return CultureInfo.CurrentCulture.TextInfo.ToTitleCase (Name.Replace ("-", " "));
			case "parent": return Parent;
			case "siblings": return Siblings;
			case "navigation": return Tacker.Navigation;
			case "ancestors": return Ancestors;
			case "current": return ctx != null && ctx.Page == this;
			}
			try {
				return Variables [key];
			} catch (KeyNotFoundException) {
				try {
					return Tacker.Metadata [key];
				} catch (KeyNotFoundException) {
					return null;
				}
			}
		}
	}
}

