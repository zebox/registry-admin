import * as React from "react";
import {
    ShowButton,
    DeleteButton,
    List,
    Datagrid,
    TextField,
    TextInput,
    useTranslate,
    useRecordContext
} from 'react-admin';

const repoFilters = [
    <TextInput source="q" label="Search" alwaysOn />,
];

const RepositoryList = () => (
    <List
        title={useTranslate()(`resources.commands.repository_name`)}
        sort={{ field: 'repository_name', order: 'ASC' }}
        perPage={25}
        filters={repoFilters}
    >
        <Datagrid bulkActionButtons={false} >
            <TextField source="id" />
            <TextField source="repository_name" />
            <SizeField source="size" />
            <ShowButton />
            <DeleteButton />
        </Datagrid>
    </List>
);

const SizeField= ({ source }: any) => {
    const record = useRecordContext();

    const convertSize=(bytes:any,decimals:number=2):string=> {
        if (!+bytes) return '0 Bytes'

        const k = 1024
        const dm = decimals < 0 ? 0 : decimals
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

        const i = Math.floor(Math.log(bytes) / Math.log(k))

        return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
    }

    return record ? (
        <>
            {convertSize(record[source],2)}
        </>
    ) : null;
}
export default RepositoryList;