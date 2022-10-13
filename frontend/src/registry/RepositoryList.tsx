import * as React from "react";
import {
    BooleanField,
    List,
    Datagrid,
    TextField,
    TextInput,
    EditButton,
    DeleteButton,
    ReferenceField,
    useTranslate

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
        </Datagrid>
    </List>
);


export default RepositoryList;