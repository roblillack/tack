using System;
using System.Collections;
using System.Collections.Generic;

namespace Tack
{
	public class DictWrapper : IDictionary<string, object>
	{
		DataProvider dataProvider;

		public DictWrapper (DataProvider provider)
		{
			dataProvider = provider;
		}

		public static DictWrapper Wrap (Page page)
		{
			return new DictWrapper (page);
		}

		public static IList<DictWrapper> Wrap (IList<Page> pages) 
		{
			var wrappers = new List<DictWrapper> ();
			foreach (var i in pages) {
				wrappers.Add (new DictWrapper (i));
			}
			return wrappers;
		}

		public static ISet<DictWrapper> Wrap (ISet<Page> pages)
		{
			var wrappers = new HashSet<DictWrapper> ();
			foreach (var i in pages) {
				wrappers.Add (new DictWrapper (i));
			}
			return wrappers;
		}

		public static IEnumerable<DictWrapper> Wrap (IEnumerable<Page> pages)
		{
			foreach (var i in pages) {
				yield return new DictWrapper (i);
			}
		}

		public void Add (string key, object value) {}
		public void Add (KeyValuePair<string, object> kvp) {}
		public bool Remove (string key) { return false; }
		public bool Remove (KeyValuePair<string, object> kvp) { return false; }
		public void Clear () {}
		public void CopyTo (KeyValuePair<string, object>[] kvps, int index) {}

		public bool ContainsKey (string key)
		{
			return dataProvider.GetData (key) != null;
		}

		public bool Contains (KeyValuePair<string, object> kvp)
		{
			return kvp.Value.Equals (dataProvider.GetData (kvp.Key));
		}

		public object this [string key] {
			get {
				return dataProvider.GetData (key);
			}
			set {}
		}

		public bool TryGetValue (string key, out object dest)
		{
			dest = dataProvider.GetData (key);
			return dest != null;
		}


		public IEnumerator<KeyValuePair<string, object>> GetEnumerator ()
		{
			return null;
		}

		System.Collections.IEnumerator System.Collections.IEnumerable.GetEnumerator ()
		{
			return null;
		}

		public int Count {
			get {
				return 0;
			}
		}

		public ICollection<string> Keys {
			get {
				return null;
			}
		}

		public ICollection<object> Values {
			get {
				return null;
			}
		}

		public bool IsReadOnly {
			get {
				return true;
			}
		}
	}
}

