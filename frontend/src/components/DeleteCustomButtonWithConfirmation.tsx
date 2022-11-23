import { DeleteWithConfirmButton, useRecordContext } from "react-admin";

export const DeleteCustomButtonWithConfirmation = (props: any) => {
    const { source } = props;
    const record = useRecordContext();

    return <DeleteWithConfirmButton {...props}
        translateOptions={{ name: record[source],id:'' }}
    />
} 