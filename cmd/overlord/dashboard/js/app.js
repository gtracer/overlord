const Clusters = {
    template: `<div>
 <div class="custom-margin">
    <h1>Clusters</h1>
    <table class="table table-striped table-sm table-hover">
        <thead>
            <tr>
                <th>Name</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="cluster in clusters">
                <td>
                    <router-link class="nav-link" :to="{ name: 'overlordminions', params:{cluster: cluster}}">{{cluster}}</router-link>
                </td >
            </tr >
        </tbody >
    </table >
</div ></div > `,
    data: function () {
        return {
            clusters: {},
        }
    },
    created: function () {
        this.fetchClusters();
        setInterval(this.fetchClusters, 10000)
    },
    methods: {
        fetchClusters: function () {
            this.$http.get("/clusters").then(result => {
                this.clusters = result.body;
            }, error => {
                console.error(error);
            });
        },
    },
}

const Minons = {
    template: `< div >
    <div class="custom-margin">
    <div class="row">
        <div class="col col-lg-2"><h1>Minions</h1></div>
        <div class="col"><button type="button" onclick="window.open('http://51.143.57.200:8090')" class="btn btn-outline-dark float-right mb-2"><i class="fa fa-tachometer" aria-hidden="true"></i>  Dashboard</button></div>
        <div class="col col-sm-2"><button @click="fetchKubeconfig" type="button" class="btn btn-outline-dark float-right mb-4"><i class="fa fa-download" aria-hidden="true"></i>  Kubeconfig</button></div>
    </div>
        <table class="table table-striped table-sm">
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Message</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="minion in minions">
                    <td>{{ minion.name }}</td>
                    <td>{{ minion.role }}</td>
                    <td><span v-bind:class="{'fa':true, 'fa-circle':true, 'green':(minion.status === 'Healthy'), 'yellow':(minion.status === 'Warning'), 'red':(minion.status === 'Unhealthy')}"></span>  {{ minion.status }}</td>
                    <td>{{ minion.message }}</td>
                </tr>
            </tbody>
        </table>
    </div></div > `,
    data: function () {
        return {
            minions: {},
        }
    },
    created: function () {
        this.fetchMinions();
        setInterval(this.fetchMinions, 10000)
    },
    methods: {
        forceFileDownload(response) {
            const url = window.URL.createObjectURL(new Blob([response]))
            const link = document.createElement('a')
            link.href = url
            link.setAttribute('download', 'kube.conf') //or any other extension
            document.body.appendChild(link)
            link.click()
        },
        fetchMinions: function () {
            this.$http.get("/" + this.$route.params.cluster + "/minions").then(result => {
                this.minions = result.body;
            }, error => {
                console.error(error);
            });
        },
        fetchKubeconfig: function (event) {
            this.$http.get("/" + this.$route.params.cluster + "/kubeconfig").then(result => {
                this.forceFileDownload(result.body)
            }, error => {
                console.error(error);
            });
        },
    },
}

const routes = [
    { path: '/', component: Clusters },
    { path: '/overlordclusters', component: Clusters },
    { path: '/:cluster/overlordminions', name: 'overlordminions', component: Minons },
]

const router = new VueRouter({
    routes // short for `routes: routes`
})

const app = new Vue({
    router
}).$mount('#app')