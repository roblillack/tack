using System;
using System.Collections.Generic;
using System.IO;
using YamlDotNet.RepresentationModel;

namespace Tack
{
	public class Page
	{
		public string Name { get; protected set; }
		public string DiskPath { get; protected set; }
		public string Permalink { get; protected set; }
		public Tacker Tacker { get; protected set; }

		string template;
		IDictionary<string, object> variables;

		public Page (Tacker tacker, string realpath)
		{
			Tacker = tacker;
			DiskPath = realpath;
			Permalink = realpath.Replace (Tacker.ContentDir, "");
			Name = Path.GetFileName (realpath);
		}

		public string Template {
			get {
				if (template != null) {
					return template;
				}
				LoadVariables ();
				if (template == null) {
					throw new FileNotFoundException ("No Template found for page " + Permalink);
				}
				return template;
			}
		}

		public IDictionary<string, object> Variables {
			get {
				return variables ?? (variables = LoadVariables ());
			}
		}

		private IDictionary<string, object> LoadVariables ()
		{
			var map = new Dictionary<string, object> ();

			foreach (var i in Directory.GetFiles (DiskPath, "*.yml")) {
				template = template ?? Path.GetFileNameWithoutExtension (i);
				var stream = new YamlStream ();
				stream.Load (new StreamReader (i));

				foreach (var doc in stream.Documents) {
					if (doc.RootNode is YamlMappingNode) {
						var seq = doc.RootNode as YamlMappingNode;
						foreach (var node in seq.Children) {
							map.Add (node.Key.ToString (), node.Value);
						}
					}
				}
			}

			return map;
		}
	}
}

