import { AuthProvider, fetchUtils } from 'react-admin';
import { BASE_URL, API_AUTH } from "./constants";


const apiAuthUrl: string = `${BASE_URL}${API_AUTH}`;
const httpClient = fetchUtils.fetchJson;

const authProvider: AuthProvider = {

    login: ({ username, password }) => {

        const options: fetchUtils.Options = {
            method: 'POST',
            mode: "cors",
            credentials: "include",
            body: JSON.stringify({
                user: username,
                passwd: password
            }),
            headers: new Headers({ 'Content-Type': 'application/json' }),
        };

        return httpClient(`${apiAuthUrl}/local/login`, options).then(({ status, json }) => {

            if (status === 200) {

                return Promise.resolve(json);
            }

            return Promise.reject();
        });

    },
    logout: () => {
        const options: fetchUtils.Options =  {
            method: 'GET',
            mode: "cors",
            credentials: "include",
            headers: new Headers({ 'Content-Type': 'application/json' }),
        };

        return httpClient(`${apiAuthUrl}/logout`, options).then(({ status }) => {

            if (status === 200) {

               return Promise.resolve();

            }
            return Promise.reject();
        }).catch(error=>{
            console.error(error);
            return Promise.reject({ redirectTo: '/login' });
        });
    },
    checkError: (error) => {
        const status = error.status;
        if (status === 401 || status === 403) {
            return Promise.reject();
        }
        return Promise.resolve();
    },
    checkAuth: () => {

        const options: fetchUtils.Options =  {
            method: 'GET',
            mode: "cors",
            credentials: "include",
            headers: new Headers({ 'Content-Type': 'application/json' }),
        };

        return httpClient(`${apiAuthUrl}/user`, options).then(({ status, json }) => {

            if (status === 401 || status === 403) {
               return Promise.reject();
            }
            if (json.error) {
               return Promise.reject();
            }
            return Promise.resolve(json);
        }).catch(error=>{

            console.error(error)
            return Promise.reject({ redirectTo: '/login' });
        });
    },
    getPermissions: () => {
        
        const options: fetchUtils.Options =  {
            method: 'GET',
            mode: "cors",
            credentials: "include",
            headers: new Headers({ 'Content-Type': 'application/json' }),
        };

        return httpClient(`${apiAuthUrl}/user`, options).then(({ status, json }) => {

            if (status === 401 || status === 403) {
               return Promise.reject();
            }
            if (json.error) {
               return Promise.reject();
            }
            return Promise.resolve(json);
        }).catch(error=>{
            console.error(error)
            return Promise.reject();
        });
    },
    getIdentity: () =>{
    const options: fetchUtils.Options =  {
        method: 'GET',
        mode: "cors",
        credentials: "include",
        headers: new Headers({ 'Content-Type': 'application/json' }),
    };

    return httpClient(`${apiAuthUrl}/user`, options).then(({ status, json }) => {

        if (status === 401 || status === 403) {
           return Promise.reject();
        }
        if (json.error) {
           return Promise.reject();
        }

        return Promise.resolve({
            id: json.id,
            fullName: json.name,
            avatar: json.picture
        })
    });
    }
};

export default authProvider;
