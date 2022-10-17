import { ReactElement } from 'react';
import { TextInput, useTranslate, useRecordContext } from 'react-admin';

export const SearchFieldTranslated = (additionalComponent?: ReactElement<any, any>[]): any => {
    const translate = useTranslate();
    let filters = [
        <TextInput source="q" label={translate('ra.action.search')} alwaysOn />
    ]
    if (additionalComponent && additionalComponent.length > 0) {
        additionalComponent.map(element => {
            filters.push(element);
        })
    }
    return filters;

}

export const ConvertUnixTimeToDate = (unixTimestamp: number): string => {
    var date = new Date(unixTimestamp * 1000);
    var year = "0" + date.getFullYear();
    var month = "0" + (date.getMonth() + 1);
    var day = "0" + date.getDate();
    var hours = "0" + date.getHours();
    var minutes = "0" + date.getMinutes()
    var seconds = "0" + date.getSeconds()


    return +day.slice(-2) +
        "/" + month.slice(-2) +
        "/" + year.slice(-2) +
        " " + hours.slice(-2) +
        ":" + minutes.slice(-2) +
        ":" + seconds.slice(-2);


}

export const ParseSizeToReadable=(bytes:any,decimals:number=2):string=> {
    if (!+bytes) return '0 Bytes'

    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}

