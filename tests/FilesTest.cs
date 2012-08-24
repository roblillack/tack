using System;
using System.Collections.Generic;
using NUnit.Framework;
using Tack.Utils;

namespace Tack.Tests
{
	[TestFixture()]
	public class FilesTest
	{
		[Test()]
		public void TestNonExistingDirEnumerators ()
		{
			foreach (var i in Files.EnumerateAllSubdirs ("/totally/non/existing/path")) {
				Assert.Fail ();
			}
		}

		[Test()]
		public void TestEnumeratingAllSubDirs ()
		{
			Assert.IsNotEmpty (new List<string> (Files.EnumerateAllSubdirs ("/usr")));
		}
	}
}

