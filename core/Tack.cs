using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using System.Text;
using System.Reflection;
using Tack.Plugins;

namespace Tack
{
	public class Tack
	{
		List<Command> commands = new List<Command> ();
		
		public ICollection<Command> Commands {
			get { return commands; }
		}

		public static bool IsUNIX {
			get {
				return Environment.OSVersion.Platform == PlatformID.Unix ||
					   Environment.OSVersion.Platform == PlatformID.MacOSX;
			}
		}
		
		public static string HomeDir {
			get {
				return IsUNIX ?
					   Environment.GetEnvironmentVariable ("HOME") :
					   Environment.GetFolderPath (Environment.SpecialFolder.UserProfile);
			}
		}
		
		public string BaseDir {
			get {
				return Path.Combine (HomeDir, (IsUNIX ? "." : "_") + "tackrc");
			}
		}
				
		public void RegisterCommands (Assembly assembly)
		{
			foreach (Type t in assembly.GetTypes ()) {
				if (!typeof(Command).IsAssignableFrom (t)) {
					continue;
				}
				
				var ctor = t.GetConstructor (System.Type.EmptyTypes);
				if (ctor == null) {
					continue;
				}
				
				var c = ctor.Invoke (null) as Command;
				if (c == null) {
					continue;
				}

				c.Application = this;
				commands.Add (c);
			}
		}
		
		public void Execute (string cmd, IList<string> parameters) {
			foreach (var c in commands) {
				if (cmd.Equals (c.Name.ToLower ())) {
					c.Execute (parameters);
					return;
				}
			}
			
			Console.Error.WriteLine ("Unknown command: {0}", cmd);
		}

		public static void Main (string[] args)
		{
			var tack = new Tack ();
			var cmd = args.Length > 0 ? (args [0]).ToLower () : "tack";
			var parameters = new List<string> (args);
			if (parameters.Count > 0) {
				parameters.RemoveAt (0);
			}
			
			tack.RegisterCommands (Assembly.GetExecutingAssembly ());
			tack.Execute(cmd, parameters);
		}
	}
}
