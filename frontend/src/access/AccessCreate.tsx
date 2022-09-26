
import {AutocompleteInput, Create, TextInput, SimpleForm, ReferenceInput, BooleanInput, SelectInput, useTranslate } from 'react-admin';


interface IActionList {
    id: string;
    name: string;
};

export const ActionList: Array<IActionList> = [
    { id: 'push', name: 'push' },
    { id: 'pull', name: 'pull' }
];

export const AccessCreate = () => {
    const translate = useTranslate();
    return (
        <Create title={translate('resources.accesses.add_title')} >
            <SimpleForm>
                <TextInput label={translate('resources.accesses.fields.name')} source="name" />
                <ReferenceInput source="owner_id" reference="users">
                    <SelectInput label={translate('resources.accesses.fields.owner_id')} emptyValue={null} emptyText='' optionText="name" optionValue="id" />
                </ReferenceInput>
                <SelectInput
                    label={translate('resources.accesses.fields.resource_type')}
                    source="acttype"
                    defaultValue={"repository"}
                    emptyValue={null}
                    choices={[{id:"repository",name:"repository"}]} />
                {/* <TextInput label={translate('resources.accesses.fields.resource_name')} source="resource_name" /> */}
                <AutocompleteInput source="category"
                label={translate('resources.accesses.fields.resource_name')}
                choices={[
                    { id: 'programming', name: 'Programming' },
                    { id: 'lifestyle', name: 'Lifestyle' },
                    { id: 'photography', name: 'Photography' },
                ]} />
                <SelectInput
                    label={translate('resources.accesses.fields.action')}
                    source="action"
                    defaultValue={"pull"}
                    emptyValue={null}
                    choices={ActionList} />
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>

        </Create>
    )
};

export default AccessCreate;