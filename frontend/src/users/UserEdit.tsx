import * as React from "react";
import {
    Edit,
    PasswordInput,
    TextInput,
    BooleanInput,
    ReferenceInput,
    SelectInput,
    SimpleForm,
    useTranslate,
    usePermissions,
    NotFound,
    required
} from 'react-admin';
import { requirePermission } from "../helpers/Helpers";
import { RoleList } from "./UsersList";

const passwordLengthEditUsers = (min: number, message: string = 'ra.validation.minLength') =>
    (value: string) => {

        // empty password field need for skip password value update when a user edit
        if (value.length===0) {
            return undefined;
        }
        return value && value.length < min ? { message, args: { min } } : undefined;
    }


const UserEdit = (props: any) => {
    const translate = useTranslate();
    const { source, ...rest } = props;
    const { permissions } = usePermissions();
    return (requirePermission(permissions, 'admin') ?
        <Edit title={translate('resources.users.edit_title')}>
            <SimpleForm>
                <TextInput label="ID" source="id" disabled />
                <TextInput label={translate('resources.users.fields.login')} source="login" disabled />
                <TextInput label={translate('resources.users.fields.name')} source="name" autoComplete='off' {...rest} validate={required()} />
                <PasswordInput label={translate('resources.users.fields.password')}
                    source="password"
                    autoComplete="new-password"
                    validate={passwordLengthEditUsers(6)}
                />
                <ReferenceInput source="group" reference="groups">
                    <SelectInput label={translate('resources.users.fields.group')}
                        emptyValue={''}
                        emptyText=''
                        optionText="name"
                        optionValue="id"
                        {...rest} validate={required()}
                    />
                </ReferenceInput>
                <SelectInput
                    label={translate('resources.users.fields.role')}
                    source="role"
                    defaultValue={"user"}
                    emptyValue={''}
                    choices={RoleList}
                    {...rest} validate={required()} />
                <BooleanInput label={translate('resources.users.fields.blocked')} source="blocked" />
                <TextInput label={translate('resources.users.fields.description')} source="description"
                    autoComplete='off' fullWidth />
            </SimpleForm>
        </Edit> : <NotFound />
    )
};

export default UserEdit;