import { useI18n } from '@/hooks/web/useI18n'
import { PlaceholderModel, FormSchema, ComponentNameEnum, ColProps } from '../types'
import { isFunction } from '@/utils/is'
import { firstUpperCase, humpToDash } from '@/utils'
import { set, get } from 'lodash-es'

const { t } = useI18n()

export const setTextPlaceholder = (schema: FormSchema): PlaceholderModel => {
  const textMap = [
    ComponentNameEnum.INPUT,
    ComponentNameEnum.AUTOCOMPLETE,
    ComponentNameEnum.INPUT_NUMBER,
    ComponentNameEnum.INPUT_PASSWORD
  ]
  const selectMap = [
    ComponentNameEnum.SELECT,
    ComponentNameEnum.TIME_PICKER,
    ComponentNameEnum.DATE_PICKER,
    ComponentNameEnum.TIME_SELECT,
    ComponentNameEnum.SELECT_V2
  ]
  if (textMap.includes(schema?.component as ComponentNameEnum)) {
    return {
      placeholder: t('common.inputText')
    }
  }
  if (selectMap.includes(schema?.component as ComponentNameEnum)) {
    const twoTextMap = ['datetimerange', 'daterange', 'monthrange', 'datetimerange', 'daterange']
    if (
      twoTextMap.includes(
        ((schema?.componentProps as any)?.type ||
          (schema?.componentProps as any)?.isRange) as string
      )
    ) {
      return {
        startPlaceholder: t('common.startTimeText'),
        endPlaceholder: t('common.endTimeText'),
        rangeSeparator: '-'
      }
    } else {
      return {
        placeholder: t('common.selectText')
      }
    }
  }
  return {}
}

export const setGridProp = (col: ColProps = {}): ColProps => {
  const colProps: ColProps = {
    ...(col.span
      ? {}
      : {
          xs: 24,
          sm: 12,
          md: 12,
          lg: 12,
          xl: 12
        }),
    ...col
  }
  return colProps
}

export const setComponentProps = (item: FormSchema): Recordable => {
  // const notNeedClearable = ['ColorPicker']

  const onEvents = (item?.componentProps as any)?.on || {}
  const newOnEvents: Recordable = {}

  for (const key in onEvents) {
    if (onEvents[key]) {
      newOnEvents[`on${firstUpperCase(key)}`] = (...args: any[]) => {
        onEvents[key](...args)
      }
    }
  }

  const componentProps: Recordable = {
    clearable: true,
    ...item.componentProps,
    ...newOnEvents
  }

  if (componentProps.slots) {
    delete componentProps.slots
  }
  if (componentProps.on) {
    delete componentProps.on
  }
  return componentProps
}

export const setItemComponentSlots = (slotsProps: Recordable = {}): Recordable => {
  const slotObj: Recordable = {}
  for (const key in slotsProps) {
    if (slotsProps[key]) {
      if (isFunction(slotsProps[key])) {
        slotObj[humpToDash(key)] = (...args: any[]) => {
          return slotsProps[key]?.(...args)
        }
      } else {
        slotObj[humpToDash(key)] = () => {
          return slotsProps[key]
        }
      }
    }
  }
  return slotObj
}

export const initModel = (schema: FormSchema[], formModel: Recordable) => {
  const model: Recordable = { ...formModel }
  schema.map((v) => {
    if (v.remove) {
      delete model[v.field]
    } else if (v.component !== 'Divider') {
      // const hasField = Reflect.has(model, v.field)
      const hasField = get(model, v.field)

      set(
        model,
        v.field,
        hasField !== void 0 ? get(model, v.field) : v.value !== void 0 ? v.value : undefined
      )
      // model[v.field] = hasField ? model[v.field] : v.value !== void 0 ? v.value : undefined
    }
  })
  return model
}
