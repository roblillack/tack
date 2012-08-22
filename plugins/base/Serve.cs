using System;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Text;
using System.Threading;
using Tack;
using Tack.Plugins;

namespace Tack.Plugins.Base
{
	public class Serve : Command
	{
        public override string Name {
			get { return "serve"; }
		}

		public override string Description {
			get { return "Runs a minimal HTTP server"; }
		}
		
		public override void Execute (IList<string> parameters)
		{
			var tacker = new Tacker (Directory.GetCurrentDirectory ());
			tacker.Tack ();

			HttpListener listener = new HttpListener();
			listener.Prefixes.Add ("http://*:8080/");
			listener.Start ();
			Console.WriteLine ("Serving from {0}, listening on port 8080 …", tacker.TargetDir);
			for(;;) {
				HttpListenerContext ctx = listener.GetContext();
				new Thread(new RequestHandler(tacker, ctx).ProcessRequest).Start();
			}
		}
	}

	public class RequestHandler {
		HttpListenerContext context;
		Tacker tacker;

		public RequestHandler (Tacker t, HttpListenerContext ctx)
		{
			context = ctx;
			tacker = t;
		}

		public void ProcessRequest ()
		{
			string msg = context.Request.HttpMethod + " " + context.Request.Url;
			Console.WriteLine(msg);

			foreach (var f in new string[] {
				tacker.TargetDir + context.Request.Url.AbsolutePath,
				Path.Combine (tacker.TargetDir + context.Request.Url.AbsolutePath, "index.html")
			}) {
				Console.WriteLine (f);
				if (!File.Exists (f)) {
					continue;
				}
				byte[] bytes = File.ReadAllBytes (f);
				context.Response.ContentLength64 = bytes.Length;
				context.Response.OutputStream.Write (bytes, 0, bytes.Length);
				context.Response.OutputStream.Close ();
				return;
			}
  
			StringBuilder sb = new StringBuilder();
			sb.Append("<h1>404 – File not found :(</h1>");
  
			byte[] b = Encoding.UTF8.GetBytes(sb.ToString());
			context.Response.StatusCode = 404;
			context.Response.StatusDescription = "File not found";
			context.Response.ContentLength64 = b.Length;
			context.Response.OutputStream.Write(b, 0, b.Length);
			context.Response.OutputStream.Close();
		}
	}
}

