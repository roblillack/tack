using System;
using System.Collections.Generic;
using Tack.Plugins;

namespace Tack.Plugins.Base
{
	public class Help : Command
	{
		public override string Name {
			get { return "help"; }
		}
		
		public override string Description {
			get { return "Displays this help screen"; }
		}
		
		public override void Execute (IList<string> parameters)
		{
			Console.WriteLine (@"tack.

usage: tack <verb> [parameters]

Available verbs:");
			foreach (Command i in Application.Commands) {
				Console.WriteLine ("    {0,-15} {1}", i.Name, i.Description);
			}
		}
	}

}

