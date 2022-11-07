
import * as React from "react";
import {
    List,
    Datagrid,
    ReferenceInput,
    AutocompleteInput,
    TextField,
    EditButton,
    DeleteButton,
    ReferenceField,
    SelectInput,
    useTranslate
} from 'react-admin';
import { DisabledField } from "../components/DisabledField";
import { SearchFieldTranslated } from '../helpers/Helpers'
import { ActionList } from "./AccessCreate";

const AccessList = () => {
const translate = useTranslate();

  return  <List
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        filters={SearchFieldTranslated([
        <ReferenceInput source="owner_id" reference="users" label="OWNER">
            <AutocompleteInput  optionText="name" optionValue="id"  />
        </ReferenceInput>,
        <SelectInput
        label={translate('resources.accesses.fields.action')}
        source="action"
        choices={ActionList} />
    ])}
    >
        <Datagrid bulkActionButtons={false} >
            <TextField source="name" label={translate('resources.accesses.fields.name')}/>
            <ReferenceField source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')}>
                <TextField source="name" />
            </ReferenceField>
            <TextField source="type" label={translate('resources.accesses.fields.resource_type')}/>
            <TextField source="resource_name" label={translate('resources.accesses.fields.resource_name')} />
            <TextField source="action" label={translate('resources.accesses.fields.action')} />
            <DisabledField source="disabled" label={translate('resources.accesses.fields.disabled')}/>
            <EditButton alignIcon="right" />
            <DeleteButton alignIcon="right" />
        </Datagrid>
    </List>
};


export default AccessList;