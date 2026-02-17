module.exports = function releasesPlugin(context) {
  return {
    name: 'releases-plugin',
    async loadContent() {
      let releases = [];
      try {
        const res = await fetch(
          'https://api.github.com/repos/dopejs/GoZen/releases?per_page=100',
          {headers: {'Accept': 'application/vnd.github.v3+json'}},
        );
        if (res.ok) {
          const all = await res.json();
          releases = all
            .filter((r) => !r.prerelease && !r.draft)
            .map((r) => ({
              tag: r.tag_name,
              name: r.name || r.tag_name,
              date: r.published_at,
              body: r.body || '',
              url: r.html_url,
            }));
        }
      } catch {
        // Build continues with empty releases
      }
      return releases;
    },
    async contentLoaded({content, actions}) {
      const {createData, addRoute} = actions;
      const dataPath = await createData('releases.json', JSON.stringify(content));
      addRoute({
        path: '/releases',
        component: '@site/src/components/ReleasesPage.tsx',
        modules: {releases: dataPath},
        exact: false,
      });
    },
  };
};
