using System;
using System.Collections;
using System.Collections.Generic;

namespace Tack
{
	public class DictWrapper : IDictionary<string, object>
	{
		DataProvider dataProvider;
		RenderContext renderContext;

		public DictWrapper (DataProvider provider, RenderContext ctx)
		{
			dataProvider = provider;
			renderContext = ctx;
		}

		public object Wrap (object o)
		{
			if (o is Page) {
				return Wrap (o as Page);
			} else if (o is IEnumerable<Page>) {
				return Wrap (o as IEnumerable<Page>);
			}

			return o;
		}

		public DictWrapper Wrap (Page page)
		{
			return new DictWrapper (page, renderContext);
		}

		public IList<DictWrapper> Wrap (IEnumerable<Page> pages) 
		{
			var wrappers = new List<DictWrapper> ();
			foreach (var i in pages) {
				wrappers.Add (Wrap (i));
			}
			return wrappers;
		}

		public void Add (string key, object value) {}
		public void Add (KeyValuePair<string, object> kvp) {}
		public bool Remove (string key) { return false; }
		public bool Remove (KeyValuePair<string, object> kvp) { return false; }
		public void Clear () {}
		public void CopyTo (KeyValuePair<string, object>[] kvps, int index) {}

		public bool ContainsKey (string key)
		{
			return dataProvider.GetData (key, renderContext) != null;
		}

		public bool Contains (KeyValuePair<string, object> kvp)
		{
			return kvp.Value.Equals (dataProvider.GetData (kvp.Key, renderContext));
		}

		public object this [string key] {
			get {
				return Wrap (dataProvider.GetData (key, renderContext));
			}
			set {}
		}

		public bool TryGetValue (string key, out object dest)
		{
			dest = dataProvider.GetData (key, renderContext);
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

