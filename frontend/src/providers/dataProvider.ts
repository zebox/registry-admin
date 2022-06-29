import {
    fetchUtils, HttpError, DataProvider,
    CreateParams, CreateResult,
    DeleteParams, DeleteResult, DeleteManyResult, DeleteManyParams,
    GetListParams, GetListResult, GetOneParams, GetManyParams, GetManyReferenceParams,
    UpdateParams, UpdateManyParams, UpdateResult, UpdateManyResult,
    GetOneResult, GetManyResult, GetManyReferenceResult, RaRecord
} from 'react-admin';
import { stringify } from 'query-string';

import { BASE_URL, API_BASE } from "./constants";


const apiUrl: string = `${BASE_URL}${API_BASE}`;
const httpClient = fetchUtils.fetchJson;

const dataProvider: DataProvider = {
    getOne: function <RecordType extends RaRecord = any>(resource: string, params: GetOneParams<any>): Promise<GetOneResult<RecordType>> {
        return httpClient(`${apiUrl}/${resource}/${params.id}`, createOptions("GET")).then(({ json }) => (json));
    },
    getList: function <RecordType extends RaRecord = any>(resource: string, params: GetListParams): Promise<GetListResult<RecordType> | any> {
        return new Promise((resolve, reject): Promise<GetListResult<any> | any> => {
            const { page, perPage } = params.pagination;
            const { field, order } = params.sort;
            const query = {
                sort: JSON.stringify([field, order]),
                range: JSON.stringify([(page - 1) * perPage, page * perPage - 1]),
                filter: JSON.stringify(params.filter),
            };
            const url = `${apiUrl}/${resource}?${stringify(query)}`;


            return httpClient(url, createOptions("GET")).then(({ status, json }) => {

                if (status === 200) {
                    return resolve(json);
                }
                return reject();
            }
            );
        });
    },
    getMany: function <RecordType extends RaRecord = any>(resource: string, params: GetManyParams): Promise<GetManyResult<RecordType>> {
        const query = {
            filter: JSON.stringify({ ids: params.ids }),
        };
        const url = `${apiUrl}/${resource}?${stringify(query)}`;
        return httpClient(url,createOptions("GET")).then(({ json }) => ( json ));
    },
    getManyReference: function <RecordType extends RaRecord = any>(resource: string, params: GetManyReferenceParams): Promise<GetManyReferenceResult<RecordType>> {
        const { page, perPage } = params.pagination;
        const { field, order } = params.sort;
        const query = {
            sort: JSON.stringify([field, order]),
            range: JSON.stringify([(page - 1) * perPage, page * perPage - 1]),
            filter: JSON.stringify({
                ...params.filter,
                [params.target]: params.id,
            }),
        };
        const url = `${apiUrl}/${resource}?${stringify(query)}`;
        return httpClient(url,createOptions("GET")).then(({ headers, json }) => (json));
        
    },
    update: function <RecordType extends RaRecord = any>(resource: string, params: UpdateParams<any>): Promise<UpdateResult<RecordType>> {
        return httpClient(`${apiUrl}/${resource}/${params.id}`, {
            method: 'PUT',
            body: JSON.stringify(params.data),
            mode: "cors",
            credentials: "include",
        }).then(({ json }) => ({ data: json }))
    },
    updateMany: function <RecordType extends RaRecord = any>(resource: string, params: UpdateManyParams<any>): Promise<UpdateManyResult<RecordType>> {
        const query = {
            filter: JSON.stringify({ id: params.ids}),
        };
        return httpClient(`${apiUrl}/${resource}?${stringify(query)}`, {
            method: 'PUT',
            body: JSON.stringify(params.data),
            mode: "cors",
            credentials: "include",
        }).then(({ json }) => ({ data: json }));
    },
    create: function <RecordType extends RaRecord = any>(resource: string, params: CreateParams<any>): Promise<CreateResult<RecordType>> {
       return httpClient(`${apiUrl}/${resource}`, {
            method: 'POST',
            body: JSON.stringify(params.data),
            mode: "cors",
            credentials: "include",
        }).then(({ json }) => ({
            data: { ...params.data, id: json.id },
        }))
    },
    delete: function <RecordType extends RaRecord = any>(resource: string, params: DeleteParams<RecordType>): Promise<DeleteResult<RecordType>> {
       return  httpClient(`${apiUrl}/${resource}/${params.id}`, {
            method: 'DELETE',
            mode: "cors",
            credentials: "include"
        }).then(({ json }) => ({ data: json }))

    },
    deleteMany: function <RecordType extends RaRecord = any>(resource: string, params: DeleteManyParams<RecordType>): Promise<DeleteManyResult<RecordType>> {
        const query = {
            filter: JSON.stringify({ id: params.ids}),
        };
        return httpClient(`${apiUrl}/${resource}?${stringify(query)}`, {
            method: 'DELETE',
            mode: "cors",
            credentials: "include"
        }).then(({ json }) => ({ data: json }));
    },

};

function createOptions(method: string): fetchUtils.Options {
    const options: fetchUtils.Options = {
        method: method,
        mode: "cors",
        credentials: "include",
        headers: new Headers({ 'Content-Type': 'application/json' }),
    };
    return options;
}

export default dataProvider;