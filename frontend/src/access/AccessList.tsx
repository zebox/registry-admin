
import { useState, useEffect } from 'react';
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
    useTranslate,
    usePermissions
} from 'react-admin';
import {DisabledField} from "../components/DisabledField";
import {SearchFieldTranslated, requirePermission} from '../helpers/Helpers'


const ActionList: Array<IActionList> = [
    { id: 'push', name: 'push' },
    { id: 'pull', name: 'pull' }
];


const AccessList = () => {
  const translate = useTranslate();
  const { permissions } = usePermissions();

  return  <List
        hasCreate={requirePermission(permissions,'admin')}
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        filters={SearchFieldTranslated([
        <ReferenceInput source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')}>
            <AutocompleteInput  optionText="name" optionValue="id" label={translate('resources.accesses.fields.owner_id')} />
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
            { requirePermission(permissions,'admin') ? <>
            <EditButton alignIcon="right" />
            <DeleteButton alignIcon="right" />
            </>:null}
        </Datagrid>
    </List>
};

interface IActionList {
    id: string;
    name: string;
};


export default AccessList;

